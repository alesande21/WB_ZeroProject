package repository

import (
	"WB_ZeroProject/internal/database"
	entity2 "WB_ZeroProject/internal/entity"
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type OrderRepo struct {
	dbRepo database.DBRepository
}

func NewOrderRepo(dbRepo database.DBRepository) *OrderRepo {
	return &OrderRepo{dbRepo: dbRepo}
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
	return nil, nil
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
		payment := order.Payment
		row := tx.QueryRowContext(ctx, queryPayment, payment.Transaction, payment.RequestId, payment.Currency,
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
		row = tx.QueryRowContext(ctx, queryOrder, order.OrderUid, order.TrackNumber, order.Entry, order.Locale,
			order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId,
			order.DateCreated, order.OofShard)
		err = row.Scan(&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId,
			&order.DateCreated, &order.OofShard)
		if err != nil {
			log.Printf("Ошибка в вставке данных в таблицу заказов: %v\n", err)
			return nil, err
		}

	}

	return newOrders, nil
}

func (r *OrderRepo) GetOrderById(ctx context.Context, orderId entity2.OrderId) (*entity2.Order, error) {
	query := `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey,
		       sm_id, date_created, oof_shard
		FROM orders
		WHERE id = $1
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
		WHERE transaction_id = $1
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
