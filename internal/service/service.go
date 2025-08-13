package service

import (
	"Effective_Mobile/internal/models"
	"github.com/google/uuid"

	"go.uber.org/zap"
)

// repository defines the interface for data access operations related to subscriptions.
// This interface allows the service layer to be decoupled from the concrete repository implementation,
// making it easier to test and swap out different data storage solutions.
type Subsrepository interface {
	CreateSubs(subs *models.Subscription) error
	UpdateSubs(id uuid.UUID, newSubs *models.Subscription) error
	DeleteSubs(id uuid.UUID) error
	ListSubs(filter models.SubscriptionFilter) ([]models.Subscription, error)
	GetSummary(sum *models.GetSummary) (int, error)
	GetSub(id uuid.UUID) (*models.Subscription, error)
	SubscriptionExists(id uuid.UUID) (bool, error)
}

// SubscriptionService provides business logic for managing subscriptions.
// It interacts with the repository layer to perform CRUD operations and data aggregation.
type SubscriptionService struct {
	repository Subsrepository
	log        *zap.Logger
}

// NewSubscriptionService creates and returns a new instance of SubscriptionService.
// It takes a repository implementation and a logger as dependencies.
func NewSubscriptionService(repository Subsrepository, log *zap.Logger) *SubscriptionService {
	return &SubscriptionService{repository: repository, log: log.Named("Service")}
}

// CreateSubs handles the creation of a new subscription.
// It delegates the operation to the underlying repository.
func (c *SubscriptionService) CreateSubs(subs *models.Subscription) error {
	// In a more complex scenario, additional business logic or validation
	// could be performed here before calling the repository.
	return c.repository.CreateSubs(subs)
}

// UpdateSubs handles the update of an existing subscription.
// It delegates the operation to the underlying repository.
func (c *SubscriptionService) UpdateSubs(id uuid.UUID, newSubs *models.Subscription) error {
	// Similar to CreateSubs, business rules for updates could be applied here.
	return c.repository.UpdateSubs(id, newSubs)
}

// DeleteSubs handles the deletion of a subscription by its ID.
// It delegates the operation to the underlying repository.
func (c *SubscriptionService) DeleteSubs(id uuid.UUID) error {
	return c.repository.DeleteSubs(id)
}

// GetSummary calculates the total cost of subscriptions based on the provided request criteria.
// It transforms the GetSummaryReq (which uses time.Time) into a GetSummary (which uses strings for dates)
// suitable for the repository layer, and then delegates the call.
func (c *SubscriptionService) GetSummary(req *models.GetSummaryReq) (int, error) {
	var fromStr, toStr string
	// Format the 'From' date from time.Time to string format "01-2006" if it's not a zero value.
	if !req.From.IsZero() {
		fromStr = req.From.Format("01-2006")
	}
	// Format the 'To' date from time.Time to string format "01-2006" if it's not a zero value.
	if !req.To.IsZero() {
		toStr = req.To.Format("01-2006")
	}

	// Create a GetSummary model for the repository layer.
	sum := models.GetSummary{
		From:        fromStr,
		To:          toStr,
		UserID:      req.UserID,
		ServiceName: req.ServiceName,
	}

	// Delegate the summary calculation to the repository.
	return c.repository.GetSummary(&sum)
}

// ListSubs retrieves a list of subscriptions based on the provided filter.
// It delegates the operation to the underlying repository.
func (c *SubscriptionService) ListSubs(filter models.SubscriptionFilter) ([]models.Subscription, error) {
	return c.repository.ListSubs(filter)
}

// GetSub retrieves a single subscription by its ID.
// It delegates the operation to the underlying repository.
func (c *SubscriptionService) GetSub(id uuid.UUID) (*models.Subscription, error) {
	return c.repository.GetSub(id)
}

// SubscriptionExists checks if a subscription with the given ID exists.
// It delegates the operation to the underlying repository.
func (c *SubscriptionService) SubscriptionExists(id uuid.UUID) (bool, error) {
	return c.repository.SubscriptionExists(id)
}
