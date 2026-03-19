package score

import "context"

type Repository interface{}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) HealthCheck(context.Context) error {
	_ = s.repository
	return nil
}
