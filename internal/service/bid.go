package service

import (
	e "AvitoProject/internal/entity"
	"context"
)

type BidRepo interface {
	CreateBid(ctx context.Context, newBid *e.Bid) (*e.Bid, error)
	CheckUserById(ctx context.Context, authorId e.BidAuthorId) (string, error)
	GetBidStatus(ctx context.Context, bidId e.BidId, authorId e.BidAuthorId) (string, error)
	GetBidsByUser(ctx context.Context, limit e.PaginationOffset, offset e.PaginationOffset, userId e.BidAuthorId) ([]e.Bid, error)
	GetBidById(ctx context.Context, bidID e.BidId) (*e.Bid, error)
	GetBidByTenderId(ctx context.Context, tenderId e.TenderId, limit e.PaginationLimit,
		offset e.PaginationOffset, statusBid []e.BidStatus) ([]e.Bid, error)
	GetBidByTenderIdByUser(ctx context.Context, tenderId e.TenderId, limit e.PaginationLimit,
		offset e.PaginationOffset, userId e.BidAuthorId) ([]e.Bid, error)
	GetBidByIdAndVersion(ctx context.Context, bidID e.BidId, version e.BidVersion) (*e.Bid, error)
	PutBidResponse(ctx context.Context, bidID e.BidId, version e.BidDecision) error
	PutReview(ctx context.Context, bidID e.BidId, username e.Username, feedback e.BidFeedback) error
	UpdateBid(ctx context.Context, bid *e.Bid) (*e.Bid, error)
	GetReviews(ctx context.Context, bidID []e.Bid, username e.Username) ([]e.BidReview, error)
}

type BidService struct {
	Repo BidRepo
}

func NewBidService(repo BidRepo) *BidService {
	return &BidService{Repo: repo}
}
