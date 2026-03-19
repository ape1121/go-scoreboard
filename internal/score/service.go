package score

import "context"

// Repository defines the persistence boundary for score operations.
type Repository interface{}

// Service is the application-level entrypoint for score use cases.
type Service struct {
	repository Repository
}

// NewService builds a score service with explicit dependencies.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// HealthCheck keeps the package compile-safe until score use cases are implemented.
func (s *Service) HealthCheck(context.Context) error {
	_ = s.repository
	return nil
}
