package repository

import (
	"WB_ZeroProject/internal/database"
	entity2 "WB_ZeroProject/internal/entity"
	"context"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/patrickmn/go-cache"
	"log"
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
	//errPing := r.Ping()
	//if errPing != nil {
	//	return nil, errPing
	//}
	//
	//query := `
	//		SELECT o.order_uid,
	//			FROM orders o
	//			LEFT JOIN tender_condition tc on o.id = o.tender_id
	//
	//	`
	//
	//var filters []string
	//var args []interface{}
	//argIndex := 1
	//
	//if len(filters) > 0 {
	//	query += " AND " + filters[0]
	//}
	//
	//query += fmt.Sprintf(" ORDER BY tc.name LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	//args = append(args, limit, offset)
	//
	//rows, err := r.dbRepo.Query(ctx, query, args...)
	//if err != nil {
	//	log.Printf("Ошибка выполнения запроса: %v\n", err)
	//	return nil, err
	//}
	//defer rows.Close()
	//
	//var tenders []e.Tender
	//for rows.Next() {
	//	var tender e.Tender
	//	err := rows.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status,
	//		&tender.OrganizationId, &tender.Version, &tender.CreatedAt)
	//	if err != nil {
	//		log.Printf("ошибка выполнения: %v\n", err)
	//		return nil, err
	//	}
	//	tenders = append(tenders, tender)
	//}
	//return tenders, nil

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

func (r *OrderRepo) CreateOrder(ctx context.Context, newOrders []entity2.Order) ([]entity2.Order, error) {
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
		return nil, fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, order := range newOrders {
		row := tx.QueryRowContext(ctx, queryOrder, order.OrderUid, order.TrackNumber, order.Entry, order.Locale,
			order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId,
			order.DateCreated, order.OofShard)
		err = row.Scan(&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId,
			&order.DateCreated, &order.OofShard)
		if err != nil {
			log.Printf("Ошибка в вставке данных в таблицу заказов: %v\n", err)
			return nil, err
		}
		delivery := order.Delivery

		row = tx.QueryRowContext(ctx, queryDelivery, order.OrderUid, delivery.Name, delivery.Phone, delivery.Zip,
			delivery.City, delivery.Address, delivery.Region, delivery.Email)
		err = row.Scan(&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address,
			&delivery.Region, &delivery.Email)
		if err != nil {
			log.Printf("Ошибка в вставке данных в таблицу доставка: %v\n", err)
			return nil, err
		}

		payment := order.Payment
		row = tx.QueryRowContext(ctx, queryPayment, payment.Transaction, payment.RequestId, payment.Currency,
			payment.Provider, payment.Amount, payment.PaymentDt, payment.Bank, payment.DeliveryCost, payment.GoodsTotal,
			payment.CustomFee)
		err = row.Scan(&payment.Transaction, &payment.RequestId, &payment.Currency, &payment.Provider, &payment.Amount,
			&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)
		if err != nil {
			log.Printf("Ошибка в вставке данных в таблицу платежей: %v\n", err)
			return nil, err
		}
		order.Payment = payment
		for i, item := range order.Items {
			row = tx.QueryRowContext(ctx, queryItem, item.ChrtId, order.OrderUid, item.TrackNumber, item.Price,
				item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status)
			err = row.Scan(&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
				&item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
			if err != nil {
				log.Printf("Ошибка в вставке данных в таблицу товаров: %v\n", err)
				return nil, err
			}
			order.Items[i] = item
		}

	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return newOrders, nil
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
		log.Printf("Заказ по orderId не найден: %v\n", err)
		return nil, err
	}

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
		log.Printf("Платежные данные по orderId не найдены: %v\n", err)
		return nil, err
	}

	order.Payment = payment

	query = `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_id = $1
	`

	rows, err := r.dbRepo.Query(ctx, query, orderId)
	if err != nil {
		log.Printf("Товары для заказа не найдены: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var items []entity2.Item
	for rows.Next() {
		var item entity2.Item
		err := rows.Scan(&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
		if err != nil {
			log.Printf("ошибка выполнения при обработке товаров: %v\n", err)
			return nil, err
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
		return 0, err
	}

	return count, nil
}

func (r *OrderRepo) Ping() error {
	return r.dbRepo.Ping()
}

func (r *OrderRepo) UpdateCache(ctx context.Context) {
	c, err := r.GetOrderCount(ctx)
	if err != nil || c == 0 {
		log.Printf("Ошибка обновление кеша: %v. Количество элементов в базе данных: %d", err, c)
		return
	}

	orders, err := r.GetOrders(ctx, entity2.PaginationLimit(c), 0)
	if err != nil {
		log.Printf("Ошибка при обновление кеша: %v", err)
		return
	}

	for _, order := range orders {
		r.cache.Set(order.OrderUid, order)
	}
	log.Printf("Кеш обновлен. Количество элементов в кеше/базе данных: %d/%d", r.cache.ItemCount(), c)
}

func (r *OrderRepo) ListenForDbChanges(ctx context.Context, updateCache <-chan interface{}) {
	tickerCacheCheck := time.NewTicker(time.Second * 51)
	tickerFullUpdate := time.NewTicker(6 * time.Minute)
	defer tickerCacheCheck.Stop()
	defer tickerFullUpdate.Stop()
	var isUpdating bool
	var mu sync.Mutex

	for {
		select {
		case <-ctx.Done():
			log.Println("Прекращается обновление кеша...")
			return
		case <-updateCache:
			log.Println("Получено новое соединение с базой данных, обновляем кэш...")
			mu.Lock()
			if !isUpdating {
				isUpdating = true
				mu.Unlock()
				r.UpdateCache(ctx)
				mu.Lock()
				isUpdating = false
			} else {
				log.Println("кеш уже в процессе обновления...")
			}
			mu.Unlock()
		case <-tickerCacheCheck.C:
			mu.Lock()
			log.Println("Проверка состояния кеша...")
			if !isUpdating {
				isUpdating = true
				mu.Unlock()
				c, err := r.GetOrderCount(ctx)
				if err != nil {
					log.Printf("Ошибка при проверке состояние кеша: %v\n", err)
				} else if c != r.cache.ItemCount() {
					log.Printf("Количество элементов в базе данных: %d/%d, обновляем кэш...", c, r.cache.ItemCount())
					r.UpdateCache(ctx)
				}
				mu.Lock()
				isUpdating = false
			} else {
				log.Println("кеш уже в процессе обновления...")
			}
			mu.Unlock()
		case <-tickerFullUpdate.C:
			mu.Lock()
			log.Println("Полное обновление кеша...")
			if !isUpdating {
				isUpdating = true
				mu.Unlock()
				r.UpdateCache(ctx)
				mu.Lock()
				isUpdating = false
			} else {
				log.Println("кеш уже в процессе обновления...")
			}
			mu.Unlock()
		}
	}
}

func (r *OrderRepo) GetOrderByIdFromCache(orderId entity2.OrderId) (*entity2.Order, error) {
	if r.cache.ItemCount() == 0 {
		return nil, fmt.Errorf("нет элементов в кеше")
	}

	order, found := r.cache.Get(orderId)
	if !found {
		return nil, fmt.Errorf("заказ под %s найден", orderId)
	}

	return order, nil
}
