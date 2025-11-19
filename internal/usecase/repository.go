package usecase

import (
	"amartha-test/internal/entity"
)

type LoanRepository interface {
	Save(loan *entity.Loan) error
	FindByID(id int) (*entity.Loan, error)
	Update(loan *entity.Loan) error
}

type BillingService struct {
	Repo LoanRepository
}
