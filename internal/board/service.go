package board

import "context"

// Repository defines the persistence boundary for leaderboard metadata.
type Repository interface{}

// Service is the application-level entrypoint for board use cases.
type Service struct {
	repository Repository
}

// NewService builds a board service with explicit dependencies.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// HealthCheck keeps the package compile-safe until board use cases are implemented.
func (s *Service) HealthCheck(context.Context) error {
	_ = s.repository
	return nil
}
