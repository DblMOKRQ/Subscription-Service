package handlers

import (
	"Effective_Mobile/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockSubscriptionService is a mock implementation of the service interface
type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) CreateSubs(subs *models.Subscription) error {
	args := m.Called(subs)
	return args.Error(0)
}

func (m *MockSubscriptionService) UpdateSubs(id uuid.UUID, newSubs *models.Subscription) error {
	args := m.Called(id, newSubs)
	return args.Error(0)
}

func (m *MockSubscriptionService) DeleteSubs(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSubscriptionService) ListSubs(filter models.SubscriptionFilter) ([]models.Subscription, error) {
	args := m.Called(filter)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) GetSummary(sum *models.GetSummaryReq) (int, error) {
	args := m.Called(sum)
	return args.Int(0), args.Error(1)
}

func (m *MockSubscriptionService) GetSub(id uuid.UUID) (*models.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) SubscriptionExists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func TestCreateSubs(t *testing.T) {
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	logger, _ := zap.NewDevelopment()
	ctx := context.WithValue(context.Background(), "logger", logger)

	// Test case 1: Successful creation
	subReq := models.SubReq{
		ServiceName: "Test Service",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
		EndDate:     nil,
	}
	reqBody, _ := json.Marshal(subReq)

	mockService.On("CreateSubs", mock.AnythingOfType("*models.Subscription")).Return(nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateSubs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp models.Response
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Successfully created subscription", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 2: Invalid request body (JSON decode error)
	reqBody = []byte(`invalid json`)
	req = httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.CreateSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid request body", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 3: Invalid request body (validation error)
	subReqInvalid := models.SubReq{
		ServiceName: "", // Invalid service name
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
		EndDate:     nil,
	}
	reqBody, _ = json.Marshal(subReqInvalid)
	req = httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.CreateSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Contains(t, resp.Msg, "Invalid request body: invalid service name")
	mockService.AssertExpectations(t)

	// Test case 4: Service error
	subReqValid := models.SubReq{
		ServiceName: "Another Service",
		Price:       200,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
		EndDate:     nil,
	}
	reqBody, _ = json.Marshal(subReqValid)
	mockService.On("CreateSubs", mock.AnythingOfType("*models.Subscription")).Return(errors.New("service error")).Once()
	req = httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.CreateSubs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Failed to create subscription", resp.Msg)
	mockService.AssertExpectations(t)
}

func TestGetSubs(t *testing.T) {
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	logger, _ := zap.NewDevelopment()
	ctx := context.WithValue(context.Background(), "logger", logger)

	// Test case 1: Successful retrieval
	id := uuid.New()
	expectedSub := &models.Subscription{ID: id, ServiceName: "Test Service"}
	mockService.On("GetSub", id).Return(expectedSub, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/subscriptions?id="+id.String(), nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetSubs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp models.Response
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Successfully get subscriptions", resp.Msg)
	assert.NotNil(t, resp.Data)
	mockService.AssertExpectations(t)

	// Test case 2: Missing ID parameter
	req = httptest.NewRequest(http.MethodGet, "/subscriptions", nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.GetSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Missing id parameter", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 3: Invalid ID format
	req = httptest.NewRequest(http.MethodGet, "/subscriptions?id=invalid-uuid", nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.GetSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid id format", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 4: Service error (e.g., subscription not found)
	mockService.On("GetSub", id).Return(nil, errors.New("subscription not found")).Once()
	req = httptest.NewRequest(http.MethodGet, "/subscriptions?id="+id.String(), nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.GetSubs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Failed to get subscription", resp.Msg)
	mockService.AssertExpectations(t)
}

func TestUpdSubs(t *testing.T) {
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	logger, _ := zap.NewDevelopment()
	ctx := context.WithValue(context.Background(), "logger", logger)

	id := uuid.New()
	// Test case 1: Successful update
	subReq := models.SubReq{
		ServiceName: "Updated Service",
		Price:       200,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
		EndDate:     nil,
	}
	reqBody, _ := json.Marshal(subReq)

	mockService.On("SubscriptionExists", id).Return(true, nil).Once()
	mockService.On("UpdateSubs", id, mock.AnythingOfType("*models.Subscription")).Return(nil).Once()

	req := httptest.NewRequest(http.MethodPut, "/subscriptions?id="+id.String(), bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateSubs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp models.Response
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Successfully updated subscription", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 2: Missing ID parameter
	req = httptest.NewRequest(http.MethodPut, "/subscriptions", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.UpdateSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Missing id parameter", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 3: Invalid ID format
	req = httptest.NewRequest(http.MethodPut, "/subscriptions?id=invalid-uuid", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.UpdateSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid id format, example xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 4: Subscription does not exist
	mockService.On("SubscriptionExists", id).Return(false, nil).Once()
	req = httptest.NewRequest(http.MethodPut, "/subscriptions?id="+id.String(), bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.UpdateSubs(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Subscription does not exist", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 5: Service error during update
	mockService.On("SubscriptionExists", id).Return(true, nil).Once()
	mockService.On("UpdateSubs", id, mock.AnythingOfType("*models.Subscription")).Return(errors.New("service update error")).Once()
	req = httptest.NewRequest(http.MethodPut, "/subscriptions?id="+id.String(), bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.UpdateSubs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Failed to update subscription", resp.Msg)
	mockService.AssertExpectations(t)
}

func TestDeleteSubs(t *testing.T) {
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	logger, _ := zap.NewDevelopment()
	ctx := context.WithValue(context.Background(), "logger", logger)

	// Test case 1: Successful deletion
	id := uuid.New()
	mockService.On("DeleteSubs", id).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions?id="+id.String(), nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.DeleteSubs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp models.Response
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Successfully deleted subscription", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 2: Missing ID parameter
	req = httptest.NewRequest(http.MethodDelete, "/subscriptions", nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.DeleteSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Missing id parameter", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 3: Invalid ID format
	req = httptest.NewRequest(http.MethodDelete, "/subscriptions?id=invalid-uuid", nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.DeleteSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid id format", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 4: Service error
	mockService.On("DeleteSubs", id).Return(errors.New("service delete error")).Once()
	req = httptest.NewRequest(http.MethodDelete, "/subscriptions?id="+id.String(), nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.DeleteSubs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Failed to delete subscription", resp.Msg)
	mockService.AssertExpectations(t)
}

func TestListSubs(t *testing.T) {
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	logger, _ := zap.NewDevelopment()
	ctx := context.WithValue(context.Background(), "logger", logger)

	// Test case 1: Successful listing with no filters
	filter := models.SubscriptionFilter{UserID: nil, ServiceName: nil}
	expectedSubs := []models.Subscription{{ID: uuid.New(), ServiceName: "Service A"}}
	mockService.On("ListSubs", filter).Return(expectedSubs, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/all-subscriptions", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ListSubs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp models.Response
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Successfully get list subs", resp.Msg)
	assert.NotNil(t, resp.Data)
	mockService.AssertExpectations(t)

	// Test case 2: Successful listing with userId filter
	userID := uuid.New()
	filterWithUser := models.SubscriptionFilter{UserID: &userID, ServiceName: nil}
	mockService.On("ListSubs", filterWithUser).Return(expectedSubs, nil).Once()

	req = httptest.NewRequest(http.MethodGet, "/all-subscriptions?userId="+userID.String(), nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ListSubs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Successfully get list subs", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 3: Invalid userId filter
	req = httptest.NewRequest(http.MethodGet, "/all-subscriptions?userId=invalid", nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ListSubs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid user id parameter", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 4: Service error
	mockService.On("ListSubs", filter).Return([]models.Subscription{}, errors.New("service list error")).Once()
	req = httptest.NewRequest(http.MethodGet, "/all-subscriptions", nil).WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ListSubs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Failed to get list subs", resp.Msg)
	mockService.AssertExpectations(t)
}

func TestGetSummary(t *testing.T) {
	mockService := new(MockSubscriptionService)
	handler := NewSubscriptionHandler(mockService)

	logger, _ := zap.NewDevelopment()
	ctx := context.WithValue(context.Background(), "logger", logger)

	// Test case 1: Successful summary retrieval
	sumReq := models.GetSummaryReq{
		From: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC),
	}
	reqBody, _ := json.Marshal(sumReq)

	mockService.On("GetSummary", &sumReq).Return(500, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/summary", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.GetSummary(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp models.Response
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Successfully get summary", resp.Msg)
	assert.NotNil(t, resp.Data)
	mockService.AssertExpectations(t)

	// Test case 2: Invalid request body
	reqBody = []byte(`invalid json`)
	req = httptest.NewRequest(http.MethodPost, "/subscriptions/summary", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.GetSummary(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid request body", resp.Msg)
	mockService.AssertExpectations(t)

	// Test case 3: Service error
	mockService.On("GetSummary", &sumReq).Return(0, errors.New("service summary error")).Once()
	reqBody, _ = json.Marshal(sumReq)
	req = httptest.NewRequest(http.MethodPost, "/subscriptions/summary", bytes.NewBuffer(reqBody)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.GetSummary(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Failed to get summary", resp.Msg)
	mockService.AssertExpectations(t)
}

func TestValidateSubReq(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Test case 1: Valid request
	subReq := models.SubReq{
		ServiceName: "Valid Service",
		Price:       10,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
		EndDate:     nil,
	}
	startDate, endDate, err := handler.validateSubReq(&subReq)
	assert.NoError(t, err)
	assert.Equal(t, "01-2025", startDate)
	assert.Nil(t, endDate)

	// Test case 2: Valid request with EndDate
	endDateStr := "12-2025"
	subReqWithEndDate := models.SubReq{
		ServiceName: "Valid Service",
		Price:       10,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
		EndDate:     &endDateStr,
	}
	startDate, endDate, err = handler.validateSubReq(&subReqWithEndDate)
	assert.NoError(t, err)
	assert.Equal(t, "01-2025", startDate)
	assert.Equal(t, "12-2025", *endDate)

	// Test case 3: Invalid ServiceName
	subReq.ServiceName = ""
	_, _, err = handler.validateSubReq(&subReq)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid service name")
	subReq.ServiceName = "Valid Service" // Reset

	// Test case 4: Invalid Price
	subReq.Price = 0
	_, _, err = handler.validateSubReq(&subReq)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid price")
	subReq.Price = 10 // Reset

	// Test case 5: Invalid UserID
	subReq.UserID = uuid.Nil
	_, _, err = handler.validateSubReq(&subReq)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid user id")
	subReq.UserID = uuid.New() // Reset

	// Test case 6: Invalid StartDate format
	subReq.StartDate = "2025-01"
	_, _, err = handler.validateSubReq(&subReq)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid start date")
	subReq.StartDate = "01-2025" // Reset

	// Test case 7: Invalid EndDate format
	invalidEndDateStr := "2025-12"
	subReq.EndDate = &invalidEndDateStr
	_, _, err = handler.validateSubReq(&subReq)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid end date")
}

func TestSendResponse(t *testing.T) {
	handler := &SubscriptionHandler{}

	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	message := "Test message"
	status := http.StatusOK

	handler.sendResponse(w, data, message, status)

	assert.Equal(t, status, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp models.Response
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, status, resp.Status)
	assert.Equal(t, message, resp.Msg)
	assert.Equal(t, data["key"], resp.Data.(map[string]interface{})["key"])
}
