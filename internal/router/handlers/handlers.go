package handlers

import (
	"Effective_Mobile/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type subscriptionService interface {
	CreateSubs(subs *models.Subscription) error
	UpdateSubs(id uuid.UUID, newSubs *models.Subscription) error
	DeleteSubs(id uuid.UUID) error
	ListSubs(filter models.SubscriptionFilter) ([]models.Subscription, error)
	GetSummary(sum *models.GetSummaryReq) (int, error)
	GetSub(id uuid.UUID) (*models.Subscription, error)
	SubscriptionExists(id uuid.UUID) (bool, error)
}
type SubscriptionHandler struct {
	service subscriptionService
}

// NewSubscriptionHandler creates and returns a new instance of SubscriptionHandler.
// It takes a SubscriptionService as a dependency, allowing for dependency injection.
func NewSubscriptionHandler(service subscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// CreateSubs handles the creation of a new subscription.
// @Summary Создать новую подписку
// @Description Создает новую подписку для пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.SubReq true "Данные подписки"
// @Success 200 {object} models.Response{data=models.Subscription}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubs(w http.ResponseWriter, r *http.Request) {
	// Retrieve logger from request context. This logger includes request-specific fields.
	log := r.Context().Value("logger").(*zap.Logger)

	log.Info("Handling create subscription")
	var subReq models.SubReq
	// Decode the JSON request body into a SubReq struct.
	if err := json.NewDecoder(r.Body).Decode(&subReq); err != nil {
		log.Warn("Invalid request body", zap.Error(err))
		h.sendResponse(w, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the incoming subscription request data.
	startDate, endDate, err := h.validateSubReq(&subReq)

	if err != nil {
		log.Warn("Invalid request body", zap.Error(err))
		h.sendResponse(w, nil, fmt.Sprintf("Invalid request body: %s", err), http.StatusBadRequest)
		return
	}

	// Create a new Subscription model with a generated UUID and validated dates.
	sub := &models.Subscription{
		ID:          uuid.New(), // Generate a new UUID for the subscription
		ServiceName: subReq.ServiceName,
		Price:       subReq.Price,
		UserID:      subReq.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Call the service layer to create the subscription in the database.
	if err := h.service.CreateSubs(sub); err != nil {
		log.Warn("Failed to create subscription", zap.Error(err))
		h.sendResponse(w, nil, "Failed to create subscription", http.StatusInternalServerError)
		return
	}
	log.Info("Successfully created subscription")
	// Send a success response with the created subscription data.
	h.sendResponse(w, sub, "Successfully created subscription", http.StatusOK)
}

// GetSubs handles retrieving a subscription by its ID.
// @Summary Получить подписку по ID
// @Description Возвращает подписку по её идентификатору
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id query string true "ID подписки"
// @Success 200 {object} models.Response{data=models.Subscription}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /subscriptions [get]
func (h *SubscriptionHandler) GetSubs(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value("logger").(*zap.Logger)
	log.Info("Handling get subscriptions")
	query := r.URL.Query()

	// Extract the 'id' parameter from the URL query.
	idStr := query.Get("id")
	if idStr == "" {
		log.Warn("Missing id parameter")
		h.sendResponse(w, nil, "Missing id parameter", http.StatusBadRequest)
		return
	}
	// Parse the ID string into a UUID.
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn("Invalid id parameter", zap.String("id", idStr))
		h.sendResponse(w, nil, "Invalid id format", http.StatusBadRequest)
		return
	}

	// Call the service layer to get the subscription.
	sub, err := h.service.GetSub(id)
	if err != nil {
		log.Warn("Failed to get subscription", zap.Error(err))
		// In a real application, differentiate between not found (404) and other errors (500).
		h.sendResponse(w, nil, "Failed to get subscription", http.StatusInternalServerError)
		return
	}

	log.Info("Successfully get subscription")
	// Send a success response with the retrieved subscription data.
	h.sendResponse(w, sub, "Successfully get subscriptions", http.StatusOK)

}

// UpdateSubs handles updating an existing subscription.
// @Summary Обновить подписку
// @Description Обновляет данные существующей подписки
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id query string true "ID подписки"
// @Param subscription body models.SubReq true "Новые данные подписки"
// @Success 200 {object} models.Response{data=models.Subscription}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /subscriptions [put]
func (h *SubscriptionHandler) UpdateSubs(w http.ResponseWriter, r *http.Request) {

	log := r.Context().Value("logger").(*zap.Logger)

	log.Info("Handling update subscription")

	// Extract the 'id' parameter from the URL query.
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		log.Warn("Missing id parameter")
		h.sendResponse(w, nil, "Missing id parameter", http.StatusBadRequest)
		return
	}

	// Parse the ID string into a UUID.
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn("Invalid id parameter", zap.String("id", idStr))
		h.sendResponse(w, nil, "Invalid id format, example xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", http.StatusBadRequest)
		return
	}

	var subReq models.SubReq
	// Decode the JSON request body into a SubReq struct.
	if err := json.NewDecoder(r.Body).Decode(&subReq); err != nil {
		log.Warn("Invalid request body")
		h.sendResponse(w, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the incoming subscription request data.
	startDate, endDate, err := h.validateSubReq(&subReq)

	if err != nil {
		log.Warn("Invalid request body", zap.Error(err))
		h.sendResponse(w, nil, fmt.Sprintf("Invalid request body: %s", err), http.StatusBadRequest)
		return
	}

	// Create a new Subscription model with a generated UUID and validated dates.
	// Note: The ID here is a new UUID, but the update operation uses the ID from the URL query.
	// This might be a point of confusion or potential bug if the intent was to update the existing ID.
	sub := &models.Subscription{
		ID:          uuid.New(), // This ID is not used for the update operation, as 'id' from URL is used.
		ServiceName: subReq.ServiceName,
		Price:       subReq.Price,
		UserID:      subReq.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Check if the subscription exists before attempting to update.
	exists, err := h.service.SubscriptionExists(id)
	if err != nil || !exists {
		log.Warn("Subscription does not exist", zap.Error(err))
		h.sendResponse(w, nil, "Subscription does not exist", http.StatusNotFound)
		return
	}

	// Call the service layer to update the subscription.
	if err := h.service.UpdateSubs(id, sub); err != nil {
		log.Warn("Failed to update subscription", zap.Error(err))
		h.sendResponse(w, nil, "Failed to update subscription", http.StatusInternalServerError)
		return
	}

	log.Info("Successfully updated subscription")
	// Send a success response with the updated subscription data.
	h.sendResponse(w, sub, "Successfully updated subscription", http.StatusOK)
}

// DeleteSubs handles deleting a subscription by its ID.
// @Summary Удалить подписку
// @Description Удаляет подписку по её идентификатору
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id query string true "ID подписки"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /subscriptions [delete]
func (h *SubscriptionHandler) DeleteSubs(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value("logger").(*zap.Logger)

	log.Info("Handling delete subscription")

	// Extract the 'id' parameter from the URL query.
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		log.Warn("Missing id parameter")
		h.sendResponse(w, nil, "Missing id parameter", http.StatusBadRequest)
		return
	}

	// Parse the ID string into a UUID.
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn("Invalid id parameter", zap.String("id", idStr))
		h.sendResponse(w, nil, "Invalid id format", http.StatusBadRequest)
		return
	}

	// Call the service layer to delete the subscription.
	if err := h.service.DeleteSubs(id); err != nil {
		log.Warn("Failed to delete subscription", zap.Error(err))
		h.sendResponse(w, nil, "Failed to delete subscription", http.StatusInternalServerError)
		return
	}

	log.Info("Successfully deleted subscription")
	// Send a success response.
	h.sendResponse(w, nil, "Successfully deleted subscription", http.StatusOK)
}

// ListSubs handles listing subscriptions with optional filtering by user ID and service name.
// @Summary Получить список подписок
// @Description Возвращает список подписок с возможностью фильтрации
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param userId query string false "ID пользователя для фильтрации"
// @Param serviceName query string false "Название сервиса для фильтрации"
// @Success 200 {object} models.Response{data=[]models.Subscription}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /all-subscriptions [get]
func (h *SubscriptionHandler) ListSubs(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value("logger").(*zap.Logger)

	log.Info("Handling list subs")
	query := r.URL.Query()
	filter := models.SubscriptionFilter{}

	// Extract and parse optional 'userId' and 'serviceName' query parameters.
	userIdStr := query.Get("userId")
	serviceName := query.Get("serviceName")

	if userIdStr != "" {
		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			log.Warn("Invalid user id parameter", zap.String("userId", userIdStr))
			h.sendResponse(w, nil, "Invalid user id parameter", http.StatusBadRequest)
			return
		}
		filter.UserID = &userId
	}
	if serviceName != "" {
		filter.ServiceName = &serviceName
	}

	// Call the service layer to retrieve the list of subscriptions based on the filter.
	subs, err := h.service.ListSubs(filter)
	if err != nil {
		log.Warn("Failed to list subs", zap.Error(err))
		h.sendResponse(w, nil, "Failed to get list subs", http.StatusInternalServerError)
		return
	}

	log.Info("Successfully list subs", zap.Int("count", len(subs)))
	// Send a success response with the list of subscriptions.
	h.sendResponse(w, subs, "Successfully get list subs", http.StatusOK)
}

// GetSummary handles calculating the total cost of subscriptions for a given period and filters.
// @Summary Получить суммарную стоимость
// @Description Возвращает суммарную стоимость подписок за период с фильтрацией
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param summary body models.GetSummaryReq true "Параметры выборки"
// @Success 200 {object} models.Response{data=object{total=int}}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /subscriptions/summary [post]
func (h *SubscriptionHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value("logger").(*zap.Logger)

	log.Info("Handling get summary")

	var sumReq models.GetSummaryReq
	// Decode the JSON request body into a GetSummaryReq struct.
	if err := json.NewDecoder(r.Body).Decode(&sumReq); err != nil {
		log.Warn("Invalid request body")
		h.sendResponse(w, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call the service layer to calculate the summary.
	total, err := h.service.GetSummary(&sumReq)
	if err != nil {
		log.Warn("Failed to get summary", zap.Error(err))
		h.sendResponse(w, nil, "Failed to get summary", http.StatusInternalServerError)
		return
	}
	log.Info("Successfully get summary")
	// Send a success response with the total summary.
	h.sendResponse(w, struct {
		Total int `json:"total"`
	}{Total: total}, "Successfully get summary", http.StatusOK)
}

// validateSubReq performs validation on the incoming SubReq data.
// It checks for non-empty service name, positive price, valid user ID, and correct date formats.
// Returns formatted start and end dates as strings, or an error if validation fails.
func (h *SubscriptionHandler) validateSubReq(sub *models.SubReq) (string, *string, error) {
	// Validate ServiceName
	if sub.ServiceName == "" {
		return "", nil, errors.New("invalid service name")
	}
	// Validate Price
	if sub.Price <= 0 {
		return "", nil, errors.New("invalid price")
	}
	// Validate UserID
	if sub.UserID == uuid.Nil {
		return "", nil, errors.New("invalid user id")
	}

	// Parse and validate StartDate format (MM-YYYY).
	startDate, err := time.Parse("01-2006", sub.StartDate)
	if err != nil {
		return "", nil, errors.New("invalid start date")
	}
	// Parse and validate optional EndDate format (MM-YYYY).
	if sub.EndDate != nil {
		endDate, err := time.Parse("01-2006", *sub.EndDate)
		if err != nil {
			return "", nil, errors.New("invalid end date")
		}
		// Reformat EndDate to ensure consistency, though it's already parsed.
		*sub.EndDate = endDate.Format("01-2006")
	}

	// Return formatted start date and (potentially updated) end date.
	return startDate.Format("01-2006"), sub.EndDate, nil

}

// sendResponse is a helper function to standardize HTTP JSON responses.
// It sets the Content-Type header, writes the HTTP status code, and encodes the response struct to JSON.
func (h *SubscriptionHandler) sendResponse(w http.ResponseWriter, data interface{}, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.Response{
		Status: status,
		Msg:    message,
		Data:   data,
	}

	json.NewEncoder(w).Encode(response)
}
