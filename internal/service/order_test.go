package service

import (
	"WB_ZeroProject/internal/database"
	entity2 "WB_ZeroProject/internal/entity"
	"WB_ZeroProject/internal/repository"
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateOrder(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибки '%s' не ожидалось при открытии соединения с мок базой", err)
	}
	defer db.Close()

	newConnection := func() *sql.DB {
		return db
	}

	// Инициализируем транзакцию
	mock.ExpectBegin()

	// Мок для вставки заказа
	mock.ExpectQuery("INSERT INTO orders").WithArgs(
		"uid123", "track123", "entry123", "en", "signature", "customer1", "delivery_service1",
		"shard1", 123, "time", "oof1").
		WillReturnRows(sqlmock.NewRows([]string{"order_uid", "track_number", "entry", "locale", "internal_signature",
			"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard"}).
			AddRow("uid123", "track123", "entry123", "en", "signature", "customer1", "delivery_service1",
				"shard1", 123, "time", "oof1"))

	// Мок для вставки доставки
	mock.ExpectQuery("INSERT INTO delivery").WithArgs(
		"uid123", "John Doe", "123456789", "12345", "City", "Address", "Region", "john@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"name", "phone", "zip", "city", "address", "region",
			"email"}).AddRow("John Doe", "123456789", "12345", "City", "Address", "Region",
			"john@example.com"))

	// Мок для вставки платежа
	mock.ExpectQuery("INSERT INTO payment").WithArgs(
		"trans123", "req123", "USD", "provider1", 100.50, 1234567890, "Bank", 10.00, 90.50, 0.00).
		WillReturnRows(sqlmock.NewRows([]string{"transaction_id", "request_id", "currency", "provider", "amount",
			"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee"}).AddRow("trans123", "req123",
			"USD", "provider1", 100.50, 1234567890, "Bank", 10.00, 90.50, 0.00))

	// Мок для вставки товаров
	mock.ExpectQuery("INSERT INTO items").WithArgs(
		111, "uid123", "track123", 100.0, "rid123", "Item1", 10.0, "L", 90.0, 101, "Brand1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"chrt_id", "track_number", "price", "rid", "name", "sale",
			"size", "total_price", "nm_id", "brand", "status"}).AddRow(111, "track123", 100.0,
			"rid123", "Item1", 10.0, "L", 90.0, 101, "Brand1", 1))

	// Мок для коммита транзакции
	mock.ExpectCommit()

	// Инициализируем репозиторий
	postGre, err := database.CreatePostgresRepository(newConnection)
	if err != nil {
		t.Fatalf("ошибка при инициализации репозитория: %s", err)
	}
	repo := repository.NewOrderRepo(postGre, nil)

	// Создаем заказ для теста
	newOrders := []entity2.Order{
		{
			OrderUid:          "uid123",
			TrackNumber:       "track123",
			Entry:             "entry123",
			Locale:            "en",
			InternalSignature: "signature",
			CustomerId:        "customer1",
			DeliveryService:   "delivery_service1",
			Shardkey:          "shard1",
			SmId:              123,
			DateCreated:       "time",
			OofShard:          "oof1",
			Delivery: entity2.Delivery{
				Name:    "John Doe",
				Phone:   "123456789",
				Zip:     "12345",
				City:    "City",
				Address: "Address",
				Region:  "Region",
				Email:   "john@example.com",
			},
			Payment: entity2.Payment{
				Transaction:  "trans123",
				RequestId:    "req123",
				Currency:     "USD",
				Provider:     "provider1",
				Amount:       100.50,
				PaymentDt:    1234567890,
				Bank:         "Bank",
				DeliveryCost: 10.00,
				GoodsTotal:   90.50,
				CustomFee:    0.00,
			},
			Items: []entity2.Item{
				{
					ChrtId:      111,
					TrackNumber: "track123",
					Price:       100.0,
					Rid:         "rid123",
					Name:        "Item1",
					Sale:        10.0,
					Size:        "L",
					TotalPrice:  90.0,
					NmId:        101,
					Brand:       "Brand1",
					Status:      1,
				},
			},
		},
	}

	// Вызываем тестируемый метод
	orderIds, err := repo.CreateOrder(context.TODO(), newOrders)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Len(t, orderIds, 1)
	assert.Equal(t, "uid123", orderIds[0])

	// Проверяем, что все ожидания моков выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не были удовлетворены: %s", err)
	}
}

func TestOrderByIdFromDb(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибки '%s' не ожидалось при открытии соединения с мок базой", err)
	}
	defer db.Close()

	newConnection := func() *sql.DB {
		return db
	}

	// запрос из таблицы заказов
	rowsOrder := sqlmock.NewRows([]string{"order_uid", "track_number", "entry", "locale", "internal_signature", "customer_id",
		"delivery_service", "shardkey", "sm_id", "date_created", "oof_shard"}).AddRow("uid123", "track123",
		"entry123", "en", "signature", "customer1", "delivery_service1", "shard1", 123, "time", "oof1")

	mock.ExpectQuery("SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, " +
		"delivery_service, shardkey, sm_id, date_created, oof_shard").WillReturnRows(rowsOrder)

	// запрос из таблицы платежей
	rowsPayment := sqlmock.NewRows([]string{"transaction_id", "request_id", "currency", "provider", "amount",
		"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee"}).AddRow("trans123", "req123",
		"USD", "provider1", 100.50, 1234567890, "Bank", 10.00, 90.50, 0.00)

	mock.ExpectQuery("SELECT transaction_id, request_id, currency, provider, amount, payment_dt, bank, " +
		"delivery_cost, goods_total, custom_fee").WillReturnRows(rowsPayment)

	// запрос из таблицы товаров
	rowsItems := sqlmock.NewRows([]string{"chrt_id", "track_number", "price", "rid", "name", "sale",
		"size", "total_price", "nm_id", "brand", "status"}).AddRow(111, "track123", 100.0, "rid123", "Item1",
		10.0, "L", 90.0, 101, "Brand1", 1)

	mock.ExpectQuery("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, " +
		"brand, status").WillReturnRows(rowsItems)

	// Инициализируем репозиторий
	postGre, err := database.CreatePostgresRepository(newConnection)
	if err != nil {
		t.Fatalf("ошибка при инициализации репозитория: %s", err)
	}
	repo := repository.NewOrderRepo(postGre, nil)

	// Создаем заказ для теста
	newOrders := []entity2.Order{
		{
			OrderUid:          "uid123",
			TrackNumber:       "track123",
			Entry:             "entry123",
			Locale:            "en",
			InternalSignature: "signature",
			CustomerId:        "customer1",
			DeliveryService:   "delivery_service1",
			Shardkey:          "shard1",
			SmId:              123,
			DateCreated:       "time",
			OofShard:          "oof1",
			Delivery: entity2.Delivery{
				Name:    "John Doe",
				Phone:   "123456789",
				Zip:     "12345",
				City:    "City",
				Address: "Address",
				Region:  "Region",
				Email:   "john@example.com",
			},
			Payment: entity2.Payment{
				Transaction:  "trans123",
				RequestId:    "req123",
				Currency:     "USD",
				Provider:     "provider1",
				Amount:       100.50,
				PaymentDt:    1234567890,
				Bank:         "Bank",
				DeliveryCost: 10.00,
				GoodsTotal:   90.50,
				CustomFee:    0.00,
			},
			Items: []entity2.Item{
				{
					ChrtId:      111,
					TrackNumber: "track123",
					Price:       100.0,
					Rid:         "rid123",
					Name:        "Item1",
					Sale:        10.0,
					Size:        "L",
					TotalPrice:  90.0,
					NmId:        101,
					Brand:       "Brand1",
					Status:      1,
				},
			},
		},
	}

	// Вызываем тестируемый метод
	order, err := repo.GetOrderByIdFromDb(context.TODO(), newOrders[0].OrderUid)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, "uid123", order.OrderUid)

	// Проверяем, что все ожидания моков выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ожидания не были удовлетворены: %s", err)
	}
}

//func TestCreateTender(t *testing.T) {
//	// Создаем мок базы данных
//	db, mock, err := sqlmock.New()
//	if err != nil {
//		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
//	}
//	defer db.Close()
//
//	// Используем фиксированный UUID
//	const testUUID = "test-uuid"
//	serverTime := "2024-01-01T12:00:00Z"
//
//	// Ожидаем вызов первого запроса для вставки в таблицу tender
//	tenderRows := sqlmock.NewRows([]string{"id", "status", "organization_id", "created_at"}).
//		AddRow(testUUID, "Created", "org1", serverTime)
//
//	mock.ExpectQuery(`INSERT INTO tender \(id, status, organization_id, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id, status, organization_id, created_at`).
//		WithArgs(testUUID, "Created", "org1", serverTime, serverTime).
//		WillReturnRows(tenderRows)
//
//	// Ожидаем вызов второго запроса для вставки в таблицу tender_condition
//	conditionRows := sqlmock.NewRows([]string{"name", "description", "type", "version"}).
//		AddRow("Tender Name", "Tender Description", "Service Type", 1)
//
//	mock.ExpectQuery(`INSERT INTO tender_condition \(tender_id, name, description, type, version\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING name, description, type, version`).
//		WithArgs(testUUID, "Tender Name", "Tender Description", "Service Type", 1).
//		WillReturnRows(conditionRows)
//
//	// Инициализируем репозиторий
//	postGre, err := database.CreatePostgresRepository(db)
//	if err != nil {
//		t.Fatalf("Ошибка создания репозитория: %v", err)
//	}
//	repo := repository.NewTenderRepo(postGre)
//
//	// Вызываем тестируемый метод
//	newTender := e.CreateTenderJSONBody{
//		OrganizationId: "org1",
//		Name:           "Tender Name",
//		Description:    "Tender Description",
//		ServiceType:    "Service Type",
//	}
//	ctx := context.TODO()
//	createdTender, err := repo.CreateTender(ctx, newTender)
//
//	// Проверяем результат
//	if assert.NoError(t, err) {
//		assert.NotNil(t, createdTender)
//		assert.Equal(t, testUUID, createdTender.Id)
//		assert.Equal(t, "Created", createdTender.Status)
//		assert.Equal(t, "org1", createdTender.OrganizationId)
//		assert.Equal(t, "Tender Name", createdTender.Name)
//		assert.Equal(t, "Tender Description", createdTender.Description)
//		assert.Equal(t, "Service Type", createdTender.ServiceType)
//		assert.Equal(t, 1, createdTender.Version)
//	}
//
//	// Проверяем, что все ожидания моков выполнены
//	if err := mock.ExpectationsWereMet(); err != nil {
//		t.Errorf("ожидания не были удовлетворены: %s", err)
//	}
//}
