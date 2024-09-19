package repository

import (
	"AvitoProject/internal/database"
	e "AvitoProject/internal/entity"
	"AvitoProject/internal/utils"
	"context"
	"fmt"
	"github.com/lib/pq"
	"log"
	"strings"
)

type BidRepo struct {
	dbRepo database.DBRepository
}

func NewBidRepo(dbRepo database.DBRepository) *BidRepo {
	return &BidRepo{dbRepo: dbRepo}
}

func (r *BidRepo) CreateBid(ctx context.Context, newBid *e.Bid) (*e.Bid, error) {
	query := `
		INSERT INTO bid (id, status, tender_id, author_type, author_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, status, tender_id, author_type, author_id, created_at
	`

	newUuid, _ := utils.GenerateUUIDV7()
	if newUuid == "" {
		log.Println("Ошибка: не удалось сгенерировать UUID")
		return nil, fmt.Errorf("не удалось сгенерировать UUID")
	}
	serverTime := utils.GetCurrentTimeRFC3339()

	row := r.dbRepo.QueryRow(ctx, query, newUuid, e.BidStatusCreated, newBid.TenderId, newBid.AuthorType, newBid.AuthorId, serverTime, serverTime)

	err := row.Scan(&newBid.Id, &newBid.Status, &newBid.TenderId, &newBid.AuthorType, &newBid.AuthorId, &newBid.CreatedAt)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в CreateBid запрос 1: %v\n", err)
		return nil, err
	}

	query = `
		INSERT INTO bid_condition (bid_id, name, description, version)
		VALUES ($1, $2, $3, $4)
		RETURNING name, description, version
	`

	row = r.dbRepo.QueryRow(ctx, query, newUuid, newBid.Name, newBid.Description, 1)
	err = row.Scan(&newBid.Name, &newBid.Description, &newBid.Version)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в CreateBid запрос 2: %v\n", err)
		return nil, err
	}

	return newBid, nil

}

func (r *BidRepo) GetBidsByUser(ctx context.Context, limit e.PaginationOffset, offset e.PaginationOffset, userId e.BidAuthorId) ([]e.Bid, error) {
	query := `
		SELECT b.id, bc.name, bc.description, b.status, b.tender_id, b.created_at, b.author_type, b.author_id, bc.version
			FROM bid b
			LEFT JOIN bid_condition bc on b.id = bc.bid_id
			WHERE bc.version = (
				SELECT MAX(bc2.version)
				FROM bid_condition bc2
				WHERE bc2.bid_id = b.id
				)
	`
	var filters []string
	var args []interface{}
	argIndex := 1

	filters = append(filters, fmt.Sprintf("b.author_id = $%d", argIndex))
	args = append(args, userId)
	argIndex++

	if len(filters) > 0 {
		query += " AND " + filters[0]
	}

	query += fmt.Sprintf(" ORDER BY bc.name LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.dbRepo.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var bids []e.Bid
	for rows.Next() {
		var bid e.Bid
		err := rows.Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.CreatedAt,
			&bid.AuthorType, &bid.AuthorId, &bid.Version)
		if err != nil {
			log.Printf("ошибка выполнения: %v\n", err)
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, nil
}

func (r *BidRepo) CheckUserById(ctx context.Context, authorId e.BidAuthorId) (string, error) {
	query := `
		SELECT username
		FROM employee
		WHERE id = $1
	`

	row := r.dbRepo.QueryRow(ctx, query, authorId)

	var username string
	err := row.Scan(&username)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в CheckUserById: %v\n", err)
		return "", err
	}

	return username, nil
}

func (r *BidRepo) GetBidStatus(ctx context.Context, bidId e.BidId, authorId e.BidAuthorId) (string, error) {
	query := `
		SELECT status
		FROM bid
		WHERE id = $1 AND author_id = $2;
	`

	var args []interface{}

	args = append(args, bidId, authorId)

	row := r.dbRepo.QueryRow(ctx, query, args...)

	var status string
	err := row.Scan(&status)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в CheckUserById: %v\n", err)
		return "", err
	}

	return status, nil
}

func (r *BidRepo) UpdateBid(ctx context.Context, bid *e.Bid) (*e.Bid, error) {
	query := `
		UPDATE bid
		SET status = $2, updated_at = $3
		WHERE id = $1
		RETURNING id, status, tender_id, author_type, author_id, created_at
	`

	serverTime := utils.GetCurrentTimeRFC3339()

	row := r.dbRepo.QueryRow(ctx, query, bid.Id, bid.Status, serverTime)

	err := row.Scan(&bid.Id, &bid.Status, &bid.TenderId, &bid.AuthorType, &bid.AuthorId, &bid.CreatedAt)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в UpdateBid запрос 1: %v\n", err)
		return nil, err
	}

	query = `
		INSERT INTO bid_condition (bid_id, name, description, version)
		VALUES ($1, $2, $3, $4)
		RETURNING name, description, version
	`

	bid.Version += 1

	row = r.dbRepo.QueryRow(ctx, query, bid.Id, bid.Name, bid.Description, bid.Version)
	err = row.Scan(&bid.Name, &bid.Description, &bid.Version)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в UpdateBid запрос 2: %v\n", err)
		return nil, err
	}

	return bid, nil
}

func (r *BidRepo) GetBidById(ctx context.Context, bidID e.BidId) (*e.Bid, error) {
	query := `
		SELECT b.id, bc.name, bc.description, b.status, b.tender_id, b.created_at, b.author_type, b.author_id, bc.version
			FROM bid b
			LEFT JOIN bid_condition bc on b.id = bc.bid_id
			WHERE bc.version = (
				SELECT MAX(bc2.version)
				FROM bid_condition bc2
				WHERE bc2.bid_id = b.id
				)
			AND b.id = $1
	`

	row := r.dbRepo.QueryRow(ctx, query, bidID)

	var bid e.Bid
	err := row.Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.CreatedAt,
		&bid.AuthorType, &bid.AuthorId, &bid.Version)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в GetBidById: %v\n", err)
		return nil, err
	}

	return &bid, nil
}

func (r *BidRepo) GetBidByTenderId(ctx context.Context, tenderId e.TenderId, limit e.PaginationLimit,
	offset e.PaginationOffset, statusBid []e.BidStatus) ([]e.Bid, error) {
	query := `
		SELECT b.id, bc.name, bc.description, b.status, b.tender_id, b.created_at, b.author_type, b.author_id, bc.version
			FROM bid b
			LEFT JOIN bid_condition bc on b.id = bc.bid_id
			WHERE bc.version = (
				SELECT MAX(bc2.version)
				FROM bid_condition bc2
				WHERE bc2.bid_id = b.id
				)
			AND b.tender_id = $1
	`

	var filters []string
	var args []interface{}
	argIndex := 2
	args = append(args, tenderId)

	if len(statusBid) > 0 {
		var statusTypesStr []string
		for _, st := range statusBid {
			statusTypesStr = append(statusTypesStr, string(st))
		}
		filters = append(filters, fmt.Sprintf("b.status = ANY($%d)", argIndex))
		args = append(args, pq.Array(statusTypesStr))
		argIndex++
	}

	if len(filters) > 0 {
		query += " AND " + filters[0]
	}

	query += fmt.Sprintf(" ORDER BY bc.name LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.dbRepo.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var bids []e.Bid
	for rows.Next() {
		var bid e.Bid
		err := rows.Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.CreatedAt,
			&bid.AuthorType, &bid.AuthorId, &bid.Version)
		if err != nil {
			log.Printf("ошибка выполнения: %v\n", err)
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, nil
}

func (r *BidRepo) GetBidByTenderIdByUser(ctx context.Context, tenderId e.TenderId, limit e.PaginationLimit,
	offset e.PaginationOffset, userId e.BidAuthorId) ([]e.Bid, error) {
	query := `
		SELECT b.id, bc.name, bc.description, b.status, b.tender_id, b.created_at, b.author_type, b.author_id, bc.version
		FROM bid b
				 LEFT JOIN bid_condition bc on b.id = bc.bid_id
		WHERE bc.version = (
			SELECT MAX(bc2.version)
			FROM bid_condition bc2
			WHERE bc2.bid_id = b.id
		)
		  AND b.tender_id = $1
		  AND (
				CASE 
					WHEN b.author_id = $2 THEN b.status IN ('Created', 'Published', 'Canceled')
					ELSE b.status = 'Published'
				END
			)
		ORDER BY bc.name LIMIT $3 OFFSET $4
	`

	rows, err := r.dbRepo.Query(ctx, query, tenderId, userId, limit, offset)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var bids []e.Bid
	for rows.Next() {
		var bid e.Bid
		err := rows.Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.CreatedAt,
			&bid.AuthorType, &bid.AuthorId, &bid.Version)
		if err != nil {
			log.Printf("ошибка выполнения: %v\n", err)
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, nil
}

func (r *BidRepo) GetBidByIdAndVersion(ctx context.Context, bidID e.BidId, version e.BidVersion) (*e.Bid, error) {
	query := `
		SELECT b.id, bc.name, bc.description, b.status, b.tender_id, b.created_at, b.author_type, b.author_id, bc.version
			FROM bid b
			LEFT JOIN bid_condition bc on b.id = bc.bid_id
			WHERE b.id = $1 AND bc.version = $2
	`

	row := r.dbRepo.QueryRow(ctx, query, bidID, version)

	var bid e.Bid
	err := row.Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.CreatedAt,
		&bid.AuthorType, &bid.AuthorId, &bid.Version)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в GetBidById: %v\n", err)
		return nil, err
	}

	return &bid, nil
}

func (r *BidRepo) PutBidResponse(ctx context.Context, bidID e.BidId, version e.BidDecision) error {
	query := `
		INSERT INTO bid_response (bid_id, response)
			VALUES ($1, $2)
			RETURNING bid_id
	`

	var str string
	row := r.dbRepo.QueryRow(ctx, query, bidID, version)
	err := row.Scan(&str)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в PutBidResponse: %v\n", err)
		return err
	}

	return nil
}

func (r *BidRepo) PutReview(ctx context.Context, bidID e.BidId, username e.Username, feedback e.BidFeedback) error {
	query := `
		INSERT INTO bid_feedback (bid_id, feedback, username)
			VALUES ($1, $2, $3)
			RETURNING bid_id
	`
	var str string
	row := r.dbRepo.QueryRow(ctx, query, bidID, feedback, username)
	err := row.Scan(&str)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в PutBidResponse: %v\n", err)
		return err
	}

	return nil
}

func (r *BidRepo) GetReviews(ctx context.Context, bidID []e.Bid, username e.Username) ([]e.BidReview, error) {
	bids := make([]string, len(bidID))
	args := make([]interface{}, len(bids)+1)

	for i, id := range bidID {
		bids[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id.Id
	}

	args[len(bidID)] = username

	query := fmt.Sprintf(`
		SELECT bid_id, feedback, username
			FROM bid_feedback
			WHERE bid_id IN (%s) AND username = $%d
	`, strings.Join(bids, ","), len(bidID)+1)

	rows, err := r.dbRepo.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Ошибка выполнения запроса в GetReviews: %v\n", err)
		return nil, err
	}

	var bidRevs []e.BidReview
	for rows.Next() {
		var bidRev e.BidReview
		err := rows.Scan(&bidRev.Id, &bidRev.Description, &bidRev.CreatedAt)
		if err != nil {
			log.Printf("ошибка выполнения: %v\n", err)
			return nil, err
		}
		bidRevs = append(bidRevs, bidRev)
	}

	return bidRevs, nil
}
