package repository

import (
	"WB_ZeroProject/internal/database"
	entity2 "WB_ZeroProject/internal/entity"
	"context"
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

func (r *OrderRepo) CreateOrder(ctx context.Context, newOrder entity2.CreateOrderJSONBody) (*entity2.Order, error) {
	//query := `
	//	INSERT INTO tender (id, status, organization_id, created_at, updated_at)
	//	VALUES ($1, $2, $3, $4, $5)
	//	RETURNING id, status, organization_id, created_at
	//`
	//
	//var createdTender e.Tender
	//newUuid := utils.GenerateUUID()
	//serverTime := utils.GetCurrentTimeRFC3339()
	//
	//row := r.dbRepo.QueryRow(ctx, query, newUuid, e.TenderStatusCreated, newTender.OrganizationId, serverTime, serverTime)
	//
	//err := row.Scan(&createdTender.Id, &createdTender.Status, &createdTender.OrganizationId, &createdTender.CreatedAt)
	//if err != nil {
	//	log.Printf("Ошибка выполнения запроса в CreateTender запрос 1: %v\n", err)
	//	return nil, err
	//}
	//
	//query2 := `
	//	INSERT INTO tender_condition (tender_id, name, description, type, version)
	//	VALUES ($1, $2, $3, $4, $5)
	//	RETURNING name, description, type, version
	//`
	//
	//row = r.dbRepo.QueryRow(ctx, query2, newUuid, newTender.Name, newTender.Description, newTender.ServiceType, 1)
	//err = row.Scan(&createdTender.Name, &createdTender.Description, &createdTender.ServiceType, &createdTender.Version)
	//if err != nil {
	//	log.Printf("Ошибка выполнения запроса в CreateTender запрос 2: %v\n", err)
	//	return nil, err
	//}
	//
	//return &createdTender, nil

	return nil, nil
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
		log.Printf("Ошибка выполнения запроса в GetOrderById query 1: %v\n", err)
		return nil, err
	}

	query = `
		SELECT tender_id, name, description, type, version
		FROM tender_condition
		WHERE tender_id = $1
		ORDER BY version DESC
	`

	row = r.dbRepo.QueryRow(ctx, query, tenderId)

	err = row.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Version)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в GetTenderById query 2: %v\n", err)
		return nil, err
	}

	return &tender, nil

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
