package repository

import (
	"AvitoProject/internal/database"
	e "AvitoProject/internal/entity"
	"AvitoProject/internal/utils"
	"context"
	"fmt"
	"github.com/lib/pq"
	"log"
)

type TenderRepo struct {
	dbRepo database.DBRepository
}

func NewTenderRepo(dbRepo database.DBRepository) *TenderRepo {
	return &TenderRepo{dbRepo: dbRepo}
}

/*
http://localhost:8080/tenders?limit=5&offset=0&service_type=Construction&service_type=Delivery
http://localhost:8080/tenders?limit=5&offset=0
*/

func (r *TenderRepo) GetTenders(ctx context.Context, limit e.PaginationOffset, offset e.PaginationOffset, serviceTypes []e.TenderServiceType) ([]e.Tender, error) {
	errPing := r.Ping()
	if errPing != nil {
		return nil, errPing
	}

	query := `
		SELECT t.id, tc.name, tc.description, tc.type, t.status, organization_id, tc.version, t.created_at
			FROM tender t
			LEFT JOIN tender_condition tc on t.id = tc.tender_id
			WHERE t.status = 'Published'
			  AND tc.version = (
				SELECT MAX(tc2.version)
				FROM tender_condition tc2
				WHERE tc2.tender_id = t.id
				)
	`
	var filters []string
	var args []interface{}
	argIndex := 1

	if len(serviceTypes) > 0 {
		var serviceTypesStr []string
		for _, st := range serviceTypes {
			serviceTypesStr = append(serviceTypesStr, string(st))
		}
		filters = append(filters, fmt.Sprintf("tc.type = ANY($%d)", argIndex))
		args = append(args, pq.Array(serviceTypesStr))
		argIndex++
	}

	if len(filters) > 0 {
		query += " AND " + filters[0]
	}

	query += fmt.Sprintf(" ORDER BY tc.name LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.dbRepo.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var tenders []e.Tender
	for rows.Next() {
		var tender e.Tender
		err := rows.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status,
			&tender.OrganizationId, &tender.Version, &tender.CreatedAt)
		if err != nil {
			log.Printf("ошибка выполнения: %v\n", err)
			return nil, err
		}
		tenders = append(tenders, tender)
	}

	return tenders, nil

}

func (r *TenderRepo) CheckUsername(ctx context.Context, username e.Username) (string, error) {
	query := `
		SELECT id
		FROM employee
		WHERE username = $1
	`

	row := r.dbRepo.QueryRow(ctx, query, username)

	var id string
	err := row.Scan(&id)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в CheckUsername: %v\n", err)
		return "", err
	}

	return id, nil
}

func (r *TenderRepo) CheckResponsible(ctx context.Context, orgId e.OrganizationId, userId string) error {
	query := `
		SELECT id
		FROM organization_responsible
		WHERE organization_id = $1 AND user_id = $2
	`

	row := r.dbRepo.QueryRow(ctx, query, orgId, userId)

	var id string
	err := row.Scan(&id)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в CheckResponsible: %v\n", err)
		return err
	}

	return nil
}

func (r *TenderRepo) CheckResponsibleByUser(ctx context.Context, userId string) ([]e.OrganizationId, error) {
	query := `
		SELECT organization_id
		FROM organization_responsible
		WHERE user_id = $1
	`

	rows, err := r.dbRepo.Query(ctx, query, userId)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var orgIds []e.OrganizationId
	for rows.Next() {
		var orgId e.OrganizationId
		err := rows.Scan(&orgId)
		if err != nil {
			log.Printf("ошибка выполнения: %v\n", err)
			return nil, err
		}
		orgIds = append(orgIds, orgId)
	}

	if len(orgIds) == 0 {
		return nil, nil
	}

	return orgIds, nil
}

func (r *TenderRepo) GetUserTenders(ctx context.Context, limit e.PaginationOffset, offset e.PaginationOffset,
	orgIds []e.OrganizationId) ([]e.Tender, error) {
	query := `
		SELECT t.id, tc.name, tc.description, tc.type, t.status, organization_id, tc.version, t.created_at
			FROM tender t
			LEFT JOIN tender_condition tc on t.id = tc.tender_id
			WHERE tc.version = (
				SELECT MAX(tc2.version)
				FROM tender_condition tc2
				WHERE tc2.tender_id = t.id
				)
	`
	var filters []string
	var args []interface{}
	argIndex := 1

	if len(orgIds) > 0 {
		var idsStr []string
		for _, st := range orgIds {
			idsStr = append(idsStr, st)
		}
		filters = append(filters, fmt.Sprintf("t.organization_id = ANY($%d)", argIndex))
		args = append(args, pq.Array(idsStr))
		argIndex++
	}

	if len(filters) > 0 {
		query += " AND " + filters[0]
	}

	query += fmt.Sprintf(" ORDER BY tc.name LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.dbRepo.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var tenders []e.Tender
	for rows.Next() {
		var tender e.Tender
		err := rows.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status,
			&tender.OrganizationId, &tender.Version, &tender.CreatedAt)
		if err != nil {
			log.Printf("ошибка выполнения: %v\n", err)
			return nil, err
		}
		tenders = append(tenders, tender)
	}

	return tenders, nil
}

func (r *TenderRepo) CreateTender(ctx context.Context, newTender e.CreateTenderJSONBody) (*e.Tender, error) {
	query := `
		INSERT INTO tender (id, status, organization_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, organization_id, created_at
	`

	var createdTender e.Tender
	newUuid := utils.GenerateUUID()
	serverTime := utils.GetCurrentTimeRFC3339()

	row := r.dbRepo.QueryRow(ctx, query, newUuid, e.TenderStatusCreated, newTender.OrganizationId, serverTime, serverTime)

	err := row.Scan(&createdTender.Id, &createdTender.Status, &createdTender.OrganizationId, &createdTender.CreatedAt)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в CreateTender запрос 1: %v\n", err)
		return nil, err
	}

	query2 := `
		INSERT INTO tender_condition (tender_id, name, description, type, version)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING name, description, type, version
	`

	row = r.dbRepo.QueryRow(ctx, query2, newUuid, newTender.Name, newTender.Description, newTender.ServiceType, 1)
	err = row.Scan(&createdTender.Name, &createdTender.Description, &createdTender.ServiceType, &createdTender.Version)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в CreateTender запрос 2: %v\n", err)
		return nil, err
	}

	return &createdTender, nil
}

func (r *TenderRepo) UpdateTender(ctx context.Context, tender *e.Tender) (*e.Tender, error) {
	query := `
		UPDATE tender
		SET status = $2, updated_at = $3
		WHERE id = $1
		RETURNING id, status, organization_id, updated_at
	`

	serverTime := utils.GetCurrentTimeRFC3339()

	row := r.dbRepo.QueryRow(ctx, query, tender.Id, tender.Status, serverTime)

	err := row.Scan(&tender.Id, &tender.Status, &tender.OrganizationId, &tender.CreatedAt)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в UpdateTender запрос 1: %v\n", err)
		return nil, err
	}

	query2 := `
		INSERT INTO tender_condition (tender_id, name, description, type, version)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING name, description, type, version
	`

	tender.Version += 1

	row = r.dbRepo.QueryRow(ctx, query2, tender.Id, tender.Name, tender.Description, tender.ServiceType, tender.Version)
	err = row.Scan(&tender.Name, &tender.Description, &tender.ServiceType, &tender.Version)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в UpdateTender запрос 2: %v\n", err)
		return nil, err
	}

	return tender, nil
}

func (r *TenderRepo) GetTenderById(ctx context.Context, tenderId e.TenderId) (*e.Tender, error) {
	query := `
		SELECT id, status, organization_id, created_at
		FROM tender
		WHERE id = $1
	`

	row := r.dbRepo.QueryRow(ctx, query, tenderId)

	var tender e.Tender

	err := row.Scan(&tender.Id, &tender.Status, &tender.OrganizationId, &tender.CreatedAt)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в GetTenderById query 1: %v\n", err)
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

func (r *TenderRepo) GetTenderByIdAndVersion(ctx context.Context, tenderID e.TenderId,
	version e.TenderVersion) (*e.Tender, error) {
	query := `
		SELECT t.id, tc.name, tc.description, tc.type, t.status, organization_id, tc.version, t.created_at
			FROM tender t
			LEFT JOIN tender_condition tc on t.id = tc.tender_id
			WHERE tc.version = $1 AND t.id = $2
	`

	var tender e.Tender
	row := r.dbRepo.QueryRow(ctx, query, version, tenderID)

	err := row.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status,
		&tender.OrganizationId, &tender.Version, &tender.CreatedAt)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в GetTenderById query 2: %v\n", err)
		return nil, err
	}

	return &tender, nil
}

func (r *TenderRepo) GetTenderCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM tender`

	err := r.dbRepo.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *TenderRepo) Ping() error {
	return r.dbRepo.Ping()
}
