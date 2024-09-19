package service

import (
	e "AvitoProject/internal/entity"
	"context"
)

type TenderRepo interface {
	CheckUsername(ctx context.Context, username e.Username) (string, error)
	CreateTender(ctx context.Context, newTender e.CreateTenderJSONBody) (*e.Tender, error)
	GetTenders(ctx context.Context, limit e.PaginationOffset, offset e.PaginationOffset,
		serviceTypes []e.TenderServiceType) ([]e.Tender, error)
	GetUserTenders(ctx context.Context, limit e.PaginationOffset, offset e.PaginationOffset,
		orgIds []e.OrganizationId) ([]e.Tender, error)
	CheckResponsible(ctx context.Context, orgId e.OrganizationId, userId string) error
	CheckResponsibleByUser(ctx context.Context, userId string) ([]e.OrganizationId, error)
	UpdateTender(ctx context.Context, tender *e.Tender) (*e.Tender, error)
	GetTenderById(ctx context.Context, tenderId e.TenderId) (*e.Tender, error)
	GetTenderByIdAndVersion(ctx context.Context, tenderID e.TenderId, version e.TenderVersion) (*e.Tender, error)
	GetTenderCount(ctx context.Context) (int, error)
	Ping() error
}

type TenderService struct {
	Repo TenderRepo
}

func NewTenderService(repo TenderRepo) *TenderService {
	return &TenderService{Repo: repo}
}
