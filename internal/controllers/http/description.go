package http

import (
	entity2 "AvitoProject/internal/entity"
	service2 "AvitoProject/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/invopop/yaml"
	"io"
	"log"
	"net/http"
	"os"
)

type ServerAddress struct {
	Localhost   string `json:"localhost" yaml:"localhost"`
	DefaultPort int    `json:"defaultPort" yaml:"defaultPort"`
	EnvAddress  string `env-required:"true" json:"envAddress" yaml:"envAddress" env:"SERVER_ADDRESS"`
}

func (a *ServerAddress) LoadConfigAddress(filePath string) error {
	_, err := os.Stat(filePath)
	if !(err == nil || !os.IsNotExist(err)) {
		return errors.New("конфиг для localhost и port не найден")
	}

	//if err != nil {
	//	if os.IsNotExist(err) {
	//		return errors.New("конфиг для localhost и port не найден")
	//	}
	//	return fmt.Errorf("ошибка проверки файла: %w", err)
	//}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("ошибка чтения конфига, %w", err)
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("ошибка чтения конфига, %w, %s", err, string(buf))
	}

	err = yaml.Unmarshal(buf, a)
	if err != nil {
		return fmt.Errorf("ошибка unmarshal, %w", err)
	}

	//УДАЛИТЬ
	a.Localhost = "127.0.0.1"
	a.DefaultPort = 8080
	a.EnvAddress = "127.0.0.1:8080"

	return nil
}

func (a *ServerAddress) UpdateEnvAddress() error {
	err := cleanenv.ReadEnv(a)
	if err != nil {
		return fmt.Errorf("ошибка updating env адреса сервера: %w", err)
	}
	return nil
}

type TenderServer struct {
	tenderService *service2.TenderService
	bidService    *service2.BidService
}

var _ ServerInterface = (*TenderServer)(nil)

func NewTenderServer(tenderService *service2.TenderService, bidService *service2.BidService) *TenderServer {
	return &TenderServer{
		tenderService: tenderService,
		bidService:    bidService,
	}
}

type Error struct {
	Code    int32
	Message string
}

func sendErrorResponse(w http.ResponseWriter, code int, resp entity2.ErrorResponse) {
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

func (t *TenderServer) GetUserBids(w http.ResponseWriter, r *http.Request, params entity2.GetUserBidsParams) {
	limit := params.Limit
	if limit == nil {
		var valLimit entity2.PaginationLimit = 5
		limit = &valLimit
	}

	offset := params.Offset
	if offset == nil {
		var defOffset entity2.PaginationOffset = 0
		offset = &defOffset
	}

	var username entity2.Username
	if params.Username != nil {
		username = *params.Username
	}

	if username == "" {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
	}

	userId, err := t.tenderService.Repo.CheckUsername(r.Context(), username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	bids, err := t.bidService.Repo.GetBidsByUser(r.Context(), *limit, *offset, userId)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Список бидов не найден."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bids); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})

	}

}

func (t TenderServer) CreateBid(w http.ResponseWriter, r *http.Request) {
	var contentBid entity2.CreateBidJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&contentBid); err != nil {
		log.Println("Неверный формат для предложения!")
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат для предложения."})
		return
	}
	bid := new(entity2.Bid)
	bid.TenderId = contentBid.TenderId
	bid.Description = contentBid.Description
	bid.Name = contentBid.Name
	bid.AuthorType = contentBid.AuthorType
	bid.AuthorId = contentBid.AuthorId

	//userId, err := t.tenderService.Repo.CheckUsername(r.Context(), contentBid.CreatorUsername)
	//if err != nil {
	//	sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
	//	return
	//}
	//
	//bid.AuthorId = userId
	//bid.AuthorType = entity2.User

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), bid.AuthorId)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	tender, err := t.tenderService.Repo.GetTenderById(r.Context(), contentBid.TenderId)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Тендер по такому id нет."})
		return
	}

	if tender.Status != "Published" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры. Тендера нет среди опубликованых."})
		return
	}

	if tender.OrganizationId == orgIds[0] {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	bid, err = t.bidService.Repo.CreateBid(r.Context(), bid)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bid); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}
}

func (t TenderServer) EditBid(w http.ResponseWriter, r *http.Request, bidId entity2.BidId, params entity2.EditBidParams) {
	if bidId == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Данные неправильно сформированы или не соответствуют требованиям"})
		return
	}

	var editParam entity2.EditBidJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&editParam); err != nil {
		log.Println("Данные неправильно сформированы или не соответствуют требованиям.")
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Данные неправильно сформированы или не соответствуют требованиям."})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	bid, err := t.bidService.Repo.GetBidById(r.Context(), bidId)
	if err != nil {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Предложение не найдено для текущией спецификации"})
		return
	}

	if bid.AuthorId != id {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Тендер не найден для username"})
		return
	}

	if editParam.Name != nil || editParam.Description != nil {
		if editParam.Name != nil {
			bid.Name = *editParam.Name
		}
		if editParam.Description != nil {
			bid.Description = *editParam.Description
		}

		bid, err = t.bidService.Repo.UpdateBid(r.Context(), bid)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка при обновлении статуса"})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bid); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}

}

func (t TenderServer) SubmitBidFeedback(w http.ResponseWriter, r *http.Request, bidId entity2.BidId, params entity2.SubmitBidFeedbackParams) {
	if bidId == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры."})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	bid, err := t.bidService.Repo.GetBidById(r.Context(), bidId)
	if err != nil {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Предложение не найдено для текущией спецификации"})
		return
	}

	tender, err := t.tenderService.Repo.GetTenderById(r.Context(), bid.TenderId)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Тендер не найден"})
		return
	}

	if tender.OrganizationId != orgIds[0] {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: " Недостаточно прав для выполнения действия"})
		return
	}

	err = t.bidService.Repo.PutReview(r.Context(), bidId, params.Username, params.BidFeedback)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Отзыв не поставился!"})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bid); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}

}

func (t TenderServer) RollbackBid(w http.ResponseWriter, r *http.Request, bidId entity2.BidId, version int32, params entity2.RollbackBidParams) {
	if bidId == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры."})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	bid, err := t.bidService.Repo.GetBidByIdAndVersion(r.Context(), bidId, version)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Предложение или версия не найдены."})
		return
	}

	if bid.AuthorId != id {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	bidLastVer, err := t.bidService.Repo.GetBidById(r.Context(), bidId)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Предложение или версия не найдены."})
		return
	}

	bid.Version = bidLastVer.Version

	bid, err = t.bidService.Repo.UpdateBid(r.Context(), bid)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Предложение не обновлено."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bid); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})

	}

}

func (t *TenderServer) GetBidStatus(w http.ResponseWriter, r *http.Request, bidId entity2.BidId, params entity2.GetBidStatusParams) {
	if bidId == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Данные неправильно сформированы или не соответствуют требованиям"})
		return
	}

	userId, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	status, err := t.bidService.Repo.GetBidStatus(r.Context(), bidId, userId)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Предложение не найдено."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}
}

func (t *TenderServer) UpdateBidStatus(w http.ResponseWriter, r *http.Request, bidId entity2.BidId, params entity2.UpdateBidStatusParams) {
	if bidId == "" || params.Status == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры."})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	bid, err := t.bidService.Repo.GetBidById(r.Context(), bidId)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Предложение не найдено."})
		return
	}

	if bid.AuthorId != id {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	bid.Status = params.Status
	bid, err = t.bidService.Repo.UpdateBid(r.Context(), bid)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Предложение не обновлено."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bid); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}

}

func (t *TenderServer) SubmitBidDecision(w http.ResponseWriter, r *http.Request, bidId entity2.BidId, params entity2.SubmitBidDecisionParams) {
	if bidId == "" || params.Username == "" || params.Decision == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры."})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	bid, err := t.bidService.Repo.GetBidById(r.Context(), bidId)
	if err != nil || bid.Status != "Published" {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Предложение не найдено."})
		return
	}

	tender, err := t.tenderService.Repo.GetTenderById(r.Context(), bid.TenderId)
	if err != nil || tender.Status != "Published" {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Тендер не найден."})
		return
	}

	if tender.OrganizationId != orgIds[0] {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	if params.Decision == "Approved" {
		tender.Status = "Closed"
		tender, err = t.tenderService.Repo.UpdateTender(r.Context(), tender)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка при обновлении статуса тендера."})
			return
		}
		bid.Status = "Canceled"
		bid, err = t.bidService.Repo.UpdateBid(r.Context(), bid)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка при обновлении статуса предложения."})
			return
		}
	} else if params.Decision == "Rejected" {
		bid.Status = "Canceled"
		bid, err = t.bidService.Repo.UpdateBid(r.Context(), bid)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка при обновлении статуса предложения."})
			return
		}
	}

	err = t.bidService.Repo.PutBidResponse(r.Context(), bidId, params.Decision)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка при внесении предложкения по тендеру."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bid); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}

}

/*
1. Найдем по юзернейму айди юзера
2. Посмотрим является ли он отвественным за свою организацию
3. Если является найдем все размещенные под его id размещенные им предложения и другие предложения которые не относятся к его id под статусом published
*/
func (t *TenderServer) GetBidsForTender(w http.ResponseWriter, r *http.Request, tenderId entity2.TenderId, params entity2.GetBidsForTenderParams) {
	if tenderId == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры."})
		return
	}

	limit := params.Limit
	if limit == nil {
		var valLimit entity2.PaginationLimit = 5
		limit = &valLimit
	}

	offset := params.Offset
	if offset == nil {
		var defOffset entity2.PaginationOffset = 0
		offset = &defOffset
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	_, err = t.tenderService.Repo.GetTenderById(r.Context(), tenderId)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: " Тендер или предложение не найдено."})
		return
	}

	bids, err := t.bidService.Repo.GetBidByTenderIdByUser(r.Context(), tenderId, *limit, *offset, id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: " Тендер или предложение не найдено."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bids); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}
}

func (t *TenderServer) GetBidReviews(w http.ResponseWriter, r *http.Request, tenderId entity2.TenderId, params entity2.GetBidReviewsParams) {
	if tenderId == "" || params.AuthorUsername == "" || params.RequesterUsername == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры."})
		return
	}

	limit := params.Limit
	if limit == nil {
		var valLimit entity2.PaginationLimit = 5
		limit = &valLimit
	}

	offset := params.Offset
	if offset == nil {
		var defOffset entity2.PaginationOffset = 0
		offset = &defOffset
	}

	authorId, err := t.tenderService.Repo.CheckUsername(r.Context(), params.AuthorUsername)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	requesterId, err := t.tenderService.Repo.CheckUsername(r.Context(), params.AuthorUsername)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	requesterOrgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), requesterId)
	if err != nil || len(requesterOrgIds) == 0 {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	tender, err := t.tenderService.Repo.GetTenderById(r.Context(), tenderId)
	if err != nil || tender.Status != "Published" {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Тендер не найден."})
		return
	}

	if tender.OrganizationId != requesterOrgIds[0] {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	bids, err := t.bidService.Repo.GetBidByTenderIdByUser(r.Context(), tenderId, *limit, *offset, authorId)
	if err != nil || len(bids) == 0 {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Предложение не найдено."})
		return
	}

	reviews, err := t.bidService.Repo.GetReviews(r.Context(), bids, params.AuthorUsername)
	if err != nil {
		http.Error(w, "Ошибка получения списка тендеров", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reviews); err != nil {
		http.Error(w, "Ошибка кодирования ответа", http.StatusBadRequest)
	}

}

func (t *TenderServer) CheckServer(w http.ResponseWriter, r *http.Request) {
	res := "ok"
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return
	}
}

func (t *TenderServer) GetTenders(w http.ResponseWriter, r *http.Request, params entity2.GetTendersParams) {
	limit := params.Limit
	if limit == nil {
		var valLimit entity2.PaginationLimit = 5
		if params.Offset != nil || params.ServiceType != nil {
			limit = &valLimit
		} else {
			count, err := t.tenderService.Repo.GetTenderCount(r.Context())
			if err != nil {
				sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка получения списка тендеров"})
				return
			}
			valLimit = entity2.PaginationLimit(count)
			limit = &valLimit
		}
	}

	offset := params.Offset
	if offset == nil {
		var defOffset entity2.PaginationOffset = 0
		offset = &defOffset
	}

	// Фильтр по типам услуг
	var serviceTypes []entity2.TenderServiceType
	if params.ServiceType != nil {
		serviceTypes = *params.ServiceType
	} else {
		serviceTypes = []entity2.TenderServiceType{"Delivery", "Manufacture", "Construction"}
	}

	// Получаем тендеры из репозитория
	tenders, err := t.tenderService.Repo.GetTenders(r.Context(), *limit, *offset, serviceTypes)
	if err != nil {
		http.Error(w, "Ошибка получения списка тендеров", http.StatusInternalServerError)
		return
	}

	// Возвращаем результат
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tenders); err != nil {
		http.Error(w, "Ошибка кодирования ответа", http.StatusBadRequest)
	}
}

/*
http://localhost:8080/tenders/my?offset=0&limit=5&username=asmith
*/

func (t TenderServer) GetUserTenders(w http.ResponseWriter, r *http.Request, params entity2.GetUserTendersParams) {
	limit := params.Limit
	if limit == nil {
		var valLimit entity2.PaginationLimit = 5
		limit = &valLimit
	}

	offset := params.Offset
	if offset == nil {
		var defOffset entity2.PaginationOffset = 0
		offset = &defOffset
	}

	username := params.Username
	if username == nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), *username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	tenders, err := t.tenderService.Repo.GetUserTenders(r.Context(), *limit, *offset, orgIds)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка получения тендера."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tenders); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}

}

func (t *TenderServer) CreateTender(w http.ResponseWriter, r *http.Request) {
	var newTender entity2.CreateTenderJSONBody
	if err := json.NewDecoder(r.Body).Decode(&newTender); err != nil {
		log.Println("Неверный формат для тендера!")
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат для тендера."})

		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), newTender.CreatorUsername)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	err = t.tenderService.Repo.CheckResponsible(r.Context(), newTender.OrganizationId, id)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	tender, err := t.tenderService.Repo.CreateTender(r.Context(), newTender)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка создания тендера."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tender); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}

}

/*
http://localhost:8080/api/tenders/00000000-0000-0000-0000-333333333333/edit?username=asmith
*/

func (t *TenderServer) EditTender(w http.ResponseWriter, r *http.Request, tenderId entity2.TenderId, params entity2.EditTenderParams) {
	if tenderId == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Данные неправильно сформированы или не соответствуют требованиям"})
		return
	}

	var editParam entity2.EditTenderJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&editParam); err != nil {
		log.Println("Данные неправильно сформированы или не соответствуют требованиям.")
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Данные неправильно сформированы или не соответствуют требованиям."})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	tender, err := t.tenderService.Repo.GetTenderById(r.Context(), tenderId)
	if err != nil {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Тендер не найден для текущией спецификации"})
		return
	}

	if tender.OrganizationId != orgIds[0] {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Тендер не найден для текущией спецификации"})
		return
	}

	if editParam.Name != nil || editParam.Description != nil || editParam.ServiceType != nil {
		if editParam.Name != nil {
			tender.Name = *editParam.Name
		}
		if editParam.Description != nil {
			tender.Description = *editParam.Description
		}
		if editParam.ServiceType != nil {
			tender.ServiceType = *editParam.ServiceType
		}

		tender, err = t.tenderService.Repo.UpdateTender(r.Context(), tender)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка при обновлении статуса"})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tender); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})

	}

}

func (t TenderServer) RollbackTender(w http.ResponseWriter, r *http.Request, tenderId entity2.TenderId, version int32, params entity2.RollbackTenderParams) {
	if tenderId == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры."})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	tender, err := t.tenderService.Repo.GetTenderByIdAndVersion(r.Context(), tenderId, version)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Тендер или версия не найдены."})
		return
	}

	if tender.OrganizationId != orgIds[0] {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	tenderLastVer, err := t.tenderService.Repo.GetTenderById(r.Context(), tenderId)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Тендер или версия не найдены."})
		return
	}

	tender.Version = tenderLastVer.Version

	tender, err = t.tenderService.Repo.UpdateTender(r.Context(), tender)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Тендер не обновлен."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tender); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}
}

/*
http://localhost:8080/tenders/00000000-0000-0000-0000-888888888888/status?username=asmit
*/

func (t *TenderServer) GetTenderStatus(w http.ResponseWriter, r *http.Request, tenderId entity2.TenderId, params entity2.GetTenderStatusParams) {
	if tenderId == "" || params.Username == nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Данные неправильно сформированы или не соответствуют требованиям"})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), *params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil || len(orgIds) == 0 {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	tender, err := t.tenderService.Repo.GetTenderById(r.Context(), tenderId)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Тендер не найден"})
		return
	}

	if tender.Status != "Published" && tender.OrganizationId != orgIds[0] {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tender.Status); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа"})
	}
}

func (t *TenderServer) UpdateTenderStatus(w http.ResponseWriter, r *http.Request, tenderId entity2.TenderId, params entity2.UpdateTenderStatusParams) {
	if tenderId == "" || params.Status == "" || params.Username == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Данные неправильно сформированы или не соответствуют требованиям"})
		return
	}

	id, err := t.tenderService.Repo.CheckUsername(r.Context(), params.Username)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, entity2.ErrorResponse{Reason: "Пользователь не существует или некорректен."})
		return
	}

	orgIds, err := t.tenderService.Repo.CheckResponsibleByUser(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	tender, err := t.tenderService.Repo.GetTenderById(r.Context(), tenderId)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Тендер не найден"})
		return
	}

	if tender.OrganizationId != orgIds[0] {
		sendErrorResponse(w, http.StatusForbidden, entity2.ErrorResponse{Reason: "Недостаточно прав для выполнения действия."})
		return
	}

	if tender.Status != params.Status {
		if tender.Status == "Created" || (tender.Status == "Published" && params.Status != "Created") {
			tender.Status = params.Status
			tender, err = t.tenderService.Repo.UpdateTender(r.Context(), tender)
			if err != nil {
				sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка при обновлении статуса."})
				return
			}
		} else {
			sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка при обновлении статуса."})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tender); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}

}
