package repository

import (
	"Effective_Mobile/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Repository provides methods for interacting with the PostgreSQL database.
// It encapsulates database operations related to subscriptions.
type Repository struct {
	db  *sql.DB
	log *zap.Logger
}

// NewRepository creates and returns a new instance of Repository.
// It takes a Storage (which contains the *sql.DB connection) and a logger as dependencies.
func (s *Storage) NewRepository() *Repository {
	return &Repository{db: s.db, log: s.log.Named("Repository")}
}

// CreateSubs inserts a new subscription record into the database.
// It takes a pointer to a models.Subscription struct containing the subscription data.
// Returns an error if the insertion fails.
func (r *Repository) CreateSubs(subs *models.Subscription) error {
	r.log.Debug("Creating Subscription", zap.String("userId", subs.UserID.String()))
	// SQL query to insert a new subscription.
	// Parameters are used to prevent SQL injection.
	query := `
		INSERT INTO subscriptions 
			(id, service_name, price, user_id, start_date, end_date)
		VALUES 
			($1, $2, $3, $4, $5, $6)
	`

	// Execute the SQL insert statement.
	_, err := r.db.Exec(
		query,
		subs.ID,
		subs.ServiceName,
		subs.Price,
		subs.UserID,
		subs.StartDate,
		subs.EndDate,
	)

	if err != nil {
		r.log.Error("Error creating subscription", zap.Error(err))
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	r.log.Debug("Subscription created", zap.String("userId", subs.UserID.String()))
	return nil

}

// UpdateSubs updates an existing subscription record in the database.
// It takes the ID of the subscription to update and a models.Subscription struct
// containing the new data. Returns an error if the update fails.
func (r *Repository) UpdateSubs(id uuid.UUID, newSubs *models.Subscription) error {
	r.log.Debug("Updating subscription", zap.String("id", id.String()))

	// SQL query to update an existing subscription.
	// The WHERE clause ensures that only the subscription with the specified ID is updated.
	query := `
        UPDATE subscriptions
        SET 
            service_name = $1,
            price = $2,
            start_date = $3,
            end_date = $4
        WHERE id = $5
    `

	// Execute the SQL update statement.
	_, err := r.db.Exec(
		query,
		newSubs.ServiceName,
		newSubs.Price,
		newSubs.StartDate,
		newSubs.EndDate,
		id,
	)

	if err != nil {
		r.log.Error("Error updating subscription", zap.Error(err))
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	r.log.Debug("Subscription updated", zap.String("id", id.String()))
	return nil
}

// SubscriptionExists checks if a subscription with the given ID exists in the database.
// Returns true if the subscription exists, false otherwise, and an error if the query fails.
func (r *Repository) SubscriptionExists(id uuid.UUID) (bool, error) {
	var exists bool
	// SQL query to check for the existence of a subscription by ID.
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE id = $1)`
	// Execute the query and scan the result into the 'exists' variable.
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking subscription existence: %w", err)
	}
	return exists, nil
}

// DeleteSubs deletes a subscription record from the database by its ID.
// Returns an error if the deletion fails.
func (r *Repository) DeleteSubs(id uuid.UUID) error {
	r.log.Debug("Deleting subscription", zap.String("userId", id.String()))
	// SQL query to delete a subscription.
	query := `DELETE FROM subscriptions WHERE id = $1`
	// Execute the SQL delete statement.
	_, err := r.db.Exec(query, id)
	if err != nil {
		r.log.Error("Error deleting subscription", zap.Error(err))
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	r.log.Debug("Subscription deleted", zap.String("userId", id.String()))

	return nil
}

// ListSubs retrieves a list of subscriptions from the database based on provided filters.
// It takes a models.SubscriptionFilter struct to apply optional filtering by UserID and ServiceName.
// Returns a slice of models.Subscription and an error if the query or scanning fails.
func (r *Repository) ListSubs(filter models.SubscriptionFilter) ([]models.Subscription, error) {
	r.log.Debug("Listing subscriptions")

	// SQL query to select subscriptions. The WHERE clause dynamically applies filters.
	// $1::uuid IS NULL OR user_id = $1: Filters by user_id if $1 (filter.UserID) is not NULL.
	// $2::text IS NULL OR service_name = $2: Filters by service_name if $2 (filter.ServiceName) is not NULL.
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE 
			($1::uuid IS NULL OR user_id = $1) AND
			($2::text IS NULL OR service_name = $2)
	`

	// Execute the query with the filter parameters.
	rows, err := r.db.Query(query, filter.UserID, filter.ServiceName)
	if err != nil {
		r.log.Error("Error listing subscriptions", zap.Error(err))
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close() // Ensure rows are closed after the function returns.

	var subscriptions []models.Subscription
	// Iterate over the result set and scan each row into a Subscription struct.
	for rows.Next() {
		var subs models.Subscription

		// Scan the columns into the struct fields.
		if err := rows.Scan(
			&subs.ID,
			&subs.ServiceName,
			&subs.Price,
			&subs.UserID,
			&subs.StartDate,
			&subs.EndDate,
		); err != nil {
			r.log.Error("failed to scan subscription", zap.Error(err))
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, subs)
	}
	// Check for any errors that occurred during row iteration.
	if err = rows.Err(); err != nil {
		r.log.Error("error iterating over subscription rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating over subscription rows: %w", err)
	}
	r.log.Debug("Subscription listed", zap.Int("count", len(subscriptions)))
	return subscriptions, nil
}

// GetSummary calculates the total price of subscriptions based on date range and optional filters.
// It takes a models.GetSummary struct with 'From', 'To', 'UserID', and 'ServiceName' fields.
// Returns the total sum of prices and an error if the query fails.
func (r *Repository) GetSummary(sum *models.GetSummary) (int, error) {
	r.log.Debug("Getting summary")
	// SQL query to calculate the sum of prices. COALESCE handles cases where SUM returns NULL (no matching rows).
	// The WHERE clause dynamically applies filters for date range, user ID, and service name.
	// Date comparisons use <= and >= for inclusive ranges.
	// end_date IS NULL: includes subscriptions without an end date.
	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE 
            ($1::text = '' OR start_date <= $1) AND 
            ($2::text = '' OR end_date >= $2 OR end_date IS NULL) AND
            ($3::uuid IS NULL OR user_id = $3) AND
            ($4::text = '' OR service_name = $4)
    `

	var total int
	// Execute the query and scan the result (total sum) into the 'total' variable.
	err := r.db.QueryRow(
		query,
		sum.To, // Note: The original query had $1 as 'To' and $2 as 'From'. Ensure this is intentional.
		sum.From,
		sum.UserID,
		sum.ServiceName,
	).Scan(&total)

	if err != nil {
		r.log.Error("Error getting summary", zap.Error(err))
		return 0, fmt.Errorf("failed to calculate summary: %w", err)
	}

	r.log.Debug("Summary retrieved", zap.Int("total", total))
	return total, nil
}

// GetSub retrieves a single subscription record by its ID.
// Returns a pointer to a models.Subscription struct if found, or an error if not found or a database error occurs.
func (r *Repository) GetSub(id uuid.UUID) (*models.Subscription, error) {
	r.log.Debug("Getting subscription", zap.String("userId", id.String()))
	// SQL query to select a single subscription by ID.
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date
        FROM subscriptions
        WHERE id = $1 
        LIMIT 1
    `

	var sub models.Subscription
	// Execute the query and scan the result into the Subscription struct.
	err := r.db.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
	)

	if err != nil {
		// Check if no rows were returned (subscription not found).
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Debug("Subscription not found", zap.String("userId", id.String()))
			return nil, fmt.Errorf("subscription not found")
		}
		// Handle other database errors.
		r.log.Error("Error getting subscription", zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}
	r.log.Debug("Subscription retrieved", zap.String("userId", id.String()))
	return &sub, nil
}


