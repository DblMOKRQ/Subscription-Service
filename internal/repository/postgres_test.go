package repository

import (
	"Effective_Mobile/internal/models"
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var ( // Declare variables at package level
	mockDB  *sql.DB
	sqlMock sqlmock.Sqlmock
	repo    *Repository
	logger  *zap.Logger
)

func TestMain(m *testing.M) {
	// Initialize zap logger for testing

	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// Create mock DB
	mockDB, sqlMock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	repo = &Repository{db: mockDB, log: logger.Named("TestRepository")}

	code := m.Run()

	mockDB.Close()
	os.Exit(code)
}

func TestCreateSubs(t *testing.T) {
	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: "Test Service",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
		EndDate:     nil,
	}

	sqlMock.ExpectExec("INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)").WithArgs(
		sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateSubs(sub)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test error case
	sqlMock.ExpectExec("INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)").WithArgs(
		sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	).WillReturnError(errors.New("db error"))

	err = repo.CreateSubs(sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create subscription")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdSubs(t *testing.T) {
	id := uuid.New()
	newSubs := &models.Subscription{
		ServiceName: "Updated Service",
		Price:       200,
		StartDate:   "02-2025",
		EndDate:     nil,
	}

	sqlMock.ExpectExec("UPDATE subscriptions SET service_name = $1, price = $2, start_date = $3, end_date = $4 WHERE id = $5").WithArgs(
		newSubs.ServiceName, newSubs.Price, newSubs.StartDate, newSubs.EndDate, id,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.UpdateSubs(id, newSubs)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test error case
	sqlMock.ExpectExec("UPDATE subscriptions SET service_name = $1, price = $2, start_date = $3, end_date = $4 WHERE id = $5").WithArgs(
		newSubs.ServiceName, newSubs.Price, newSubs.StartDate, newSubs.EndDate, id,
	).WillReturnError(errors.New("db error"))

	err = repo.UpdateSubs(id, newSubs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update subscription")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSubscriptionExists(t *testing.T) {
	id := uuid.New()

	// Test exists
	sqlMock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM subscriptions WHERE id = $1)").WithArgs(id).WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)
	exists, err := repo.SubscriptionExists(id)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test not exists
	sqlMock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM subscriptions WHERE id = $1)").WithArgs(id).WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(false),
	)
	exists, err = repo.SubscriptionExists(id)
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test error
	sqlMock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM subscriptions WHERE id = $1)").WithArgs(id).WillReturnError(errors.New("db error"))
	exists, err = repo.SubscriptionExists(id)
	assert.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "error checking subscription existence")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestDeleteSubs(t *testing.T) {
	id := uuid.New()

	sqlMock.ExpectExec("DELETE FROM subscriptions WHERE id = $1").WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeleteSubs(id)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test error case
	sqlMock.ExpectExec("DELETE FROM subscriptions WHERE id = $1").WithArgs(id).WillReturnError(errors.New("db error"))

	err = repo.DeleteSubs(id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete subscription")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestListSubs(t *testing.T) {
	filter := models.SubscriptionFilter{
		UserID:      nil,
		ServiceName: nil,
	}

	sub1 := models.Subscription{
		ID: uuid.New(), ServiceName: "Service A", Price: 100, UserID: uuid.New(), StartDate: "01-2025", EndDate: nil,
	}
	sub2 := models.Subscription{
		ID: uuid.New(), ServiceName: "Service B", Price: 200, UserID: uuid.New(), StartDate: "02-2025", EndDate: nil,
	}

	rows := sqlmock.NewRows([]string{"id", "service_name", "price", "user_id", "start_date", "end_date"}).
		AddRow(sub1.ID, sub1.ServiceName, sub1.Price, sub1.UserID, sub1.StartDate, sub1.EndDate).
		AddRow(sub2.ID, sub2.ServiceName, sub2.Price, sub2.UserID, sub2.StartDate, sub2.EndDate)

	sqlMock.ExpectQuery("SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE ($1::uuid IS NULL OR user_id = $1) AND ($2::text IS NULL OR service_name = $2)").WithArgs(filter.UserID, filter.ServiceName).WillReturnRows(rows)

	subs, err := repo.ListSubs(filter)
	assert.NoError(t, err)
	assert.Len(t, subs, 2)
	assert.Equal(t, sub1.ID, subs[0].ID)
	assert.Equal(t, sub2.ID, subs[1].ID)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test error case
	sqlMock.ExpectQuery("SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE ($1::uuid IS NULL OR user_id = $1) AND ($2::text IS NULL OR service_name = $2)").WithArgs(filter.UserID, filter.ServiceName).WillReturnError(errors.New("db error"))

	subs, err = repo.ListSubs(filter)
	assert.Error(t, err)
	assert.Nil(t, subs)
	assert.Contains(t, err.Error(), "failed to query subscriptions")
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test scan error
	sqlMock.ExpectQuery("SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE ($1::uuid IS NULL OR user_id = $1) AND ($2::text IS NULL OR service_name = $2)").WithArgs(filter.UserID, filter.ServiceName).WillReturnRows(
		sqlmock.NewRows([]string{"id", "service_name", "price", "user_id", "start_date", "end_date"}).AddRow("invalid-uuid", "Service C", 300, uuid.New(), "03-2025", nil),
	)
	subs, err = repo.ListSubs(filter)
	assert.Error(t, err)
	assert.Nil(t, subs)
	assert.Contains(t, err.Error(), "failed to scan subscription")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetSummary(t *testing.T) {
	sumReq := &models.GetSummary{
		From:        "01-2025",
		To:          "12-2025",
		UserID:      nil,
		ServiceName: "",
	}

	sqlMock.ExpectQuery("SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE ($1::text = '' OR start_date <= $1) AND ($2::text = '' OR end_date >= $2 OR end_date IS NULL) AND ($3::uuid IS NULL OR user_id = $3) AND ($4::text = '' OR service_name = $4)").WithArgs(
		sumReq.To, sumReq.From, sumReq.UserID, sumReq.ServiceName,
	).WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(500))

	total, err := repo.GetSummary(sumReq)
	assert.NoError(t, err)
	assert.Equal(t, 500, total)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test error case
	sqlMock.ExpectQuery("SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE ($1::text = '' OR start_date <= $1) AND ($2::text = '' OR end_date >= $2 OR end_date IS NULL) AND ($3::uuid IS NULL OR user_id = $3) AND ($4::text = '' OR service_name = $4)").WithArgs(
		sumReq.To, sumReq.From, sumReq.UserID, sumReq.ServiceName,
	).WillReturnError(errors.New("db error"))

	total, err = repo.GetSummary(sumReq)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "failed to calculate summary")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetSub(t *testing.T) {
	id := uuid.New()
	sub := models.Subscription{
		ID: id, ServiceName: "Service X", Price: 100, UserID: uuid.New(), StartDate: "01-2025", EndDate: nil,
	}

	// Test found
	rows := sqlmock.NewRows([]string{"id", "service_name", "price", "user_id", "start_date", "end_date"}).
		AddRow(sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate)
	sqlMock.ExpectQuery("SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1 LIMIT 1").WithArgs(id).WillReturnRows(rows)

	foundSub, err := repo.GetSub(id)
	assert.NoError(t, err)
	assert.Equal(t, sub.ID, foundSub.ID)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test not found
	sqlMock.ExpectQuery("SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1 LIMIT 1").WithArgs(id).WillReturnError(sql.ErrNoRows)

	foundSub, err = repo.GetSub(id)
	assert.Error(t, err)
	assert.Nil(t, foundSub)
	assert.Contains(t, err.Error(), "subscription not found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Test error
	sqlMock.ExpectQuery("SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1 LIMIT 1").WithArgs(id).WillReturnError(errors.New("db error"))

	foundSub, err = repo.GetSub(id)
	assert.Error(t, err)
	assert.Nil(t, foundSub)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
