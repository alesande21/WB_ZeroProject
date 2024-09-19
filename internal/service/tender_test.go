package service

import (
	"AvitoProject/internal/database"
	e "AvitoProject/internal/entity"
	"AvitoProject/internal/repository"
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTenders(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибки '%s' не ожидалось при открытии соединения с мок базой", err)
	}
	defer db.Close()

	// Ожидаем вызов запроса
	rows := sqlmock.NewRows([]string{"id", "name", "description", "type", "status", "organization_id", "version", "created_at"}).
		AddRow("1", "Tender 1", "Description 1", "Construction", "Published", "org1", 1, "2024-01-01T12:00:00Z")

	mock.ExpectQuery("SELECT t.id, tc.name, tc.description, tc.type, t.status").
		WillReturnRows(rows)

	// Инициализируем репозиторий
	postGre, err := database.CreatePostgresRepository(db)
	repo := repository.NewTenderRepo(postGre)

	// Вызываем тестируемый метод
	serviceTypes := []e.TenderServiceType{"Construction"}
	limit := e.PaginationLimit(5)
	offset := e.PaginationOffset(0)
	ctx := context.TODO()
	tenders, err := repo.GetTenders(ctx, limit, offset, serviceTypes)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Len(t, tenders, 1)
	assert.Equal(t, "Tender 1", tenders[0].Name)

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
