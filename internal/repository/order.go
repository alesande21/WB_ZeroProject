package repository

import (
	"WB_ZeroProject/internal/database"
	entity2 "WB_ZeroProject/internal/entity"
	"context"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/patrickmn/go-cache"
	log2 "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type OrderRepo struct {
	dbRepo database.DBRepository
	cache  *database.AllCache
}

func NewOrderRepo(dbRepo database.DBRepository, cache *database.AllCache) *OrderRepo {
	return &OrderRepo{dbRepo: dbRepo, cache: cache}
}

func (r *OrderRepo) GetOrders(ctx context.Context, limit entity2.PaginationOffset, offset entity2.PaginationOffset) ([]entity2.Order, error) {
	queryOrder := `
		SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, 
		       o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
		       p.transaction_id, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, 
		       p.delivery_cost, p.goods_total, p.custom_fee,
		       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email
		FROM orders o
		JOIN payment p ON o.order_uid = p.transaction_id
		JOIN delivery d ON o.order_uid = d.order_id
		ORDER BY o.date_created DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.dbRepo.Query(ctx, queryOrder, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса выборки заказов: %v", err)
	}
	defer rows.Close()

	var orders []entity2.Order

	for rows.Next() {
		var order entity2.Order
		var payment entity2.Payment
		var delivery entity2.Delivery

		err = rows.Scan(
			&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerId,
			&order.DeliveryService, &order.Shardkey, &order.SmId, &order.DateCreated, &order.OofShard,
			&payment.Transaction, &payment.RequestId, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt,
			&payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка в сканировании данных заказов: %v", err)
		}

		order.Payment = payment
		order.Delivery = delivery

		items, err := r.GetOrderItems(ctx, order.OrderUid)
		if err != nil {
			return nil, fmt.Errorf("ошибка при выборке товаров для заказа %s: %v", order.OrderUid, err)
		}
		order.Items = items

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по строкам: %v", err)
	}

	return orders, nil
}

func (r *OrderRepo) GetOrderItems(ctx context.Context, orderUid string) ([]entity2.Item, error) {
	queryItem := `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_id = $1
	`

	rows, err := r.dbRepo.Query(ctx, queryItem, orderUid)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса выборки товаров: %v", err)
	}
	defer rows.Close()

	var items []entity2.Item

	for rows.Next() {
		var item entity2.Item
		err = rows.Scan(&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка в сканировании данных товаров: %v", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по строкам товаров: %v", err)
	}

	return items, nil
}

func (r *OrderRepo) CreateOrder(ctx context.Context, newOrders []entity2.Order) ([]entity2.OrderId, error) {
	queryOrder := `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, 
		                   shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, 
		                   shardkey, sm_id, date_created, oof_shard
	`

	queryPayment := `
		INSERT INTO payment (transaction_id, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, 
		       goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING transaction_id, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, 
		       goods_total, custom_fee
	`

	queryItem := `
		INSERT INTO items (chrt_id, order_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
	`

	queryDelivery := `
		INSERT INTO delivery (order_id, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING name, phone, zip, city, address, region, email
	`

	tx, err := r.dbRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("-> r.dbRepo.BeginTx: не удалось начать транзакцию: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var orderIds []entity2.OrderId

	for _, order := range newOrders {
		row := tx.QueryRowContext(ctx, queryOrder, order.OrderUid, order.TrackNumber, order.Entry, order.Locale,
			order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId,
			order.DateCreated, order.OofShard)
		err = row.Scan(&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId,
			&order.DateCreated, &order.OofShard)
		if err != nil {
			return nil, fmt.Errorf("-> row.Scan: ошибка в вставке данных в таблицу заказов: %w", err)
		}

		delivery := order.Delivery

		row = tx.QueryRowContext(ctx, queryDelivery, order.OrderUid, delivery.Name, delivery.Phone, delivery.Zip,
			delivery.City, delivery.Address, delivery.Region, delivery.Email)
		err = row.Scan(&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address,
			&delivery.Region, &delivery.Email)
		if err != nil {
			return nil, fmt.Errorf("-> row.Scan: ошибка в вставке данных в таблицу доставка: %w", err)
		}

		payment := order.Payment
		row = tx.QueryRowContext(ctx, queryPayment, payment.Transaction, payment.RequestId, payment.Currency,
			payment.Provider, payment.Amount, payment.PaymentDt, payment.Bank, payment.DeliveryCost, payment.GoodsTotal,
			payment.CustomFee)
		err = row.Scan(&payment.Transaction, &payment.RequestId, &payment.Currency, &payment.Provider, &payment.Amount,
			&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)
		if err != nil {
			return nil, fmt.Errorf("-> row.Scan: ошибка в вставке данных в таблицу платежей: %w", err)
		}
		order.Payment = payment
		for i, item := range order.Items {
			row = tx.QueryRowContext(ctx, queryItem, item.ChrtId, order.OrderUid, item.TrackNumber, item.Price,
				item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status)
			err = row.Scan(&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
				&item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
			if err != nil {
				return nil, fmt.Errorf("-> row.Scan: ошибка в вставке данных в таблицу товаров: %w", err)
			}
			order.Items[i] = item
		}
		orderIds = append(orderIds, order.OrderUid)
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("-> tx.Commit: не удалось завершить транзакцию: %w", err)
	}

	for _, order := range newOrders {
		r.cache.Set(order.OrderUid, order)
	}

	return orderIds, nil
}

func (r *OrderRepo) GetOrderByIdFromDb(ctx context.Context, orderId entity2.OrderId) (*entity2.Order, error) {
	query := `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey,
		       sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1
	`

	row := r.dbRepo.QueryRow(ctx, query, orderId)

	var order entity2.Order

	err := row.Scan(&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId, &order.DateCreated, &order.OofShard)
	if err != nil {
		return nil, fmt.Errorf("-> r.dbRepo.QueryRow.Scan: заказ по orderId %s не найден: %w", orderId, err)
	}

	query = `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery
		WHERE order_id = $1
	`

	row = r.dbRepo.QueryRow(ctx, query, orderId)

	var delivery entity2.Delivery

	err = row.Scan(&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region,
		&delivery.Email)
	if err != nil {
		return nil, fmt.Errorf("-> r.dbRepo.QueryRow.Scan: доставка по orderId %s не найдена: %w", orderId, err)
	}

	order.Delivery = delivery

	query = `
		SELECT transaction_id, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, 
		       goods_total, custom_fee
		FROM payment
		WHERE transaction_id = $1
	`

	row = r.dbRepo.QueryRow(ctx, query, orderId)

	var payment entity2.Payment

	err = row.Scan(&payment.Transaction, &payment.RequestId, &payment.Currency, &payment.Provider, &payment.Amount,
		&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)
	if err != nil {
		return nil, fmt.Errorf("-> r.dbRepo.QueryRow.Scan: платежные данные по orderId %s не найдены: %w", orderId, err)
	}

	order.Payment = payment

	query = `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_id = $1
	`

	rows, err := r.dbRepo.Query(ctx, query, orderId)
	if err != nil {
		return nil, fmt.Errorf("->  r.dbRepo.Query: товары для заказа orderId %s не найдены: %w", orderId, err)
	}
	defer rows.Close()

	var items []entity2.Item
	for rows.Next() {
		var item entity2.Item
		err := rows.Scan(&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("-> rows.Next.Scan: ошибка выполнения при обработке товаров для orderId %s: %w", orderId, err)
		}
		items = append(items, item)
	}

	order.Items = items

	return &order, nil
}

func (r *OrderRepo) GetOrderCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM orders`

	err := r.dbRepo.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("-> r.dbRepo.QueryRow.Scan(): ошибка при получении количества заказов из таблицы orders: %w", err)
	}

	return count, nil
}

func (r *OrderRepo) Ping() error {
	return r.dbRepo.Ping()
}

func (r *OrderRepo) UpdateCache(ctx context.Context) {
	c, err := r.GetOrderCount(ctx)
	if err != nil || c == 0 {
		log2.Errorf("UpdateCache-> r.GetOrderCount: Количество элементов в базе данных %d. Ошибка обновление кеша: %s", c, err.Error())
		return
	}

	const batchSize = 100
	var offset int

	for {
		orders, err := r.GetOrders(ctx, entity2.PaginationLimit(batchSize), entity2.PaginationOffset(offset))
		if err != nil {
			log2.Errorf("UpdateCache-> r.GetOrders%s", err.Error())
			return
		}

		if len(orders) == 0 {
			break
		}

		for _, order := range orders {
			r.cache.Set(order.OrderUid, order)
		}

		offset += batchSize
	}

	log2.Infof("UpdateCache: кеш обновлен. Количество элементов в кеше/базе данных: %d/%d", r.cache.ItemCount(), c)
}

func (r *OrderRepo) ListenForDbChanges(ctx context.Context, updateCache <-chan interface{}) {
	tickerCacheCheck := time.NewTicker(time.Second * 51)
	tickerFullUpdate := time.NewTicker(6 * time.Minute)
	defer tickerCacheCheck.Stop()
	defer tickerFullUpdate.Stop()
	var isUpdating bool
	var mu sync.Mutex
	r.UpdateCache(ctx)
	time.Sleep(time.Second * 2)

	for {
		select {
		case <-ctx.Done():
			log2.Info("Прекращается обновление кеша...")
			return
		case <-updateCache:
			log2.Info("Получено новое соединение с базой данных, обновляем кэш...")
			mu.Lock()
			if !isUpdating {
				isUpdating = true
				mu.Unlock()
				r.UpdateCache(ctx)
				mu.Lock()
				isUpdating = false
			} else {
				log2.Info("кеш уже в процессе обновления...")
			}
			mu.Unlock()

		case <-tickerCacheCheck.C:
			mu.Lock()
			//log.Println("Проверка состояния кеша...")
			if !isUpdating {
				isUpdating = true
				mu.Unlock()
				c, err := r.GetOrderCount(ctx)
				if err != nil {
					log2.Errorf("ListenForDbChanges-> r.GetOrderCount: ошибка при проверке состояние кеша: %s", err.Error())
				} else if c != r.cache.ItemCount() {
					log2.Infof("Количество элементов в базе данных: %d/%d, обновляем кэш...", c, r.cache.ItemCount())
					r.UpdateCache(ctx)
				}
				mu.Lock()
				isUpdating = false
			} else {
				log2.Info("кеш уже в процессе обновления...")
			}
			mu.Unlock()

		case <-tickerFullUpdate.C:
			mu.Lock()
			log2.Info("Полное обновление кеша...")
			if !isUpdating {
				isUpdating = true
				mu.Unlock()
				r.UpdateCache(ctx)
				mu.Lock()
				isUpdating = false
			} else {
				log2.Info("кеш уже в процессе обновления...")
			}
			mu.Unlock()
		}
	}

}

func (r *OrderRepo) GetOrderByIdFromCache(orderId entity2.OrderId) (*entity2.Order, error) {
	if r.cache.ItemCount() == 0 {
		return nil, fmt.Errorf("-> r.cache.ItemCount(): нет элементов в кеше")
	}

	order, found := r.cache.Get(orderId)
	if !found {
		return nil, fmt.Errorf("-> r.cache.Get: заказ под id %s найден", orderId)
	}

	return order, nil
}
