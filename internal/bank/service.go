package bank

import "github.com/FelipeMCassiano/constantia/internal/domain"

type Service interface {
	RegisterUser(user *domain.User) error

	LoginUser(user *domain.User) (int, error)

	CreateTransaction(transactionRequest *domain.Transaction) error
}

type service struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &service{
		repository: r,
	}
}

func (s *service) RegisterUser(user *domain.User) error {
	return s.repository.CreateUser(user)
}

func (s *service) LoginUser(user *domain.User) (int, error) {
	return s.repository.LoginUser(user)
}

func (s *service) CreateTransaction(transaction *domain.Transaction) error {
	return s.repository.CreateTransaction(transaction)
}
