package usecase

import (
	"errors"
	"fmt"
	"math"
	"time"

	"amartha-test/internal/entity"
)

// BillingService contains the core business logic
func NewBillingService(repo LoanRepository) *BillingService {
	return &BillingService{Repo: repo}
}

// CreateLoan initializes a loan and its schedule
func (s *BillingService) CreateLoan(principal float64, rate float64, weeks int) (*entity.Loan, error) {
	totalInterest := principal * rate
	totalPayable := principal + totalInterest
	weeklyAmount := totalPayable / float64(weeks)
	startDate := time.Now()

	loan := &entity.Loan{
		Principal:     principal,
		InterestRate:  rate,
		TotalPayable:  totalPayable,
		WeeksDuration: weeks,
		StartDate:     startDate,
		Status:        entity.StatusActive,
		Installments:  make([]*entity.Installment, weeks),
	}

	// Generate Schedule
	for i := 0; i < weeks; i++ {
		dueDate := startDate.AddDate(0, 0, (i+1)*7)
		loan.Installments[i] = &entity.Installment{
			WeekNumber: i + 1,
			DueDate:    dueDate,
			AmountDue:  weeklyAmount,
			IsPaid:     false,
		}
	}

	if err := s.Repo.Save(loan); err != nil {
		return nil, err
	}
	return loan, nil
}

// GetOutstanding calculates remaining balance
func (s *BillingService) GetOutstanding(loanID int) (float64, error) {
	loan, err := s.Repo.FindByID(loanID)
	if err != nil {
		return 0, err
	}

	var outstanding float64
	for _, inst := range loan.Installments {
		if !inst.IsPaid {
			outstanding += inst.AmountDue
		}
	}
	return outstanding, nil
}

// CheckDelinquency checks status based on simulated "current date"
func (s *BillingService) CheckDelinquency(loanID int) (bool, error) {
	loan, err := s.Repo.FindByID(loanID)
	if err != nil {
		return false, err
	}

	// currentDate := time.Now() // In production, this is real time.

	// Simulate "Fast Forwarding" 5 weeks (35 days) into the future so we can test delinquency
	currentDate := time.Now().AddDate(0, 0, 35)

	missedPayments := 0
	for _, inst := range loan.Installments {
		if inst.DueDate.Before(currentDate) && !inst.IsPaid {
			missedPayments++
		}
	}

	isDelinquent := missedPayments >= 2

	// Update status in DB if changed
	if isDelinquent && loan.Status != entity.StatusDelinquent {
		loan.Status = entity.StatusDelinquent
		s.Repo.Update(loan)
	} else if !isDelinquent && loan.Status == entity.StatusDelinquent {
		loan.Status = entity.StatusActive
		s.Repo.Update(loan)
	}

	return isDelinquent, nil
}

// MakePayment handles the payment logic (FIFO)
func (s *BillingService) MakePayment(loanID int, amount float64) error {
	loan, err := s.Repo.FindByID(loanID)
	if err != nil {
		return err
	}

	// FIFO Logic: Find first unpaid installment
	var target *entity.Installment
	for _, inst := range loan.Installments {
		if !inst.IsPaid {
			target = inst
			break
		}
	}

	if target == nil {
		return errors.New("loan is already fully paid")
	}

	// Strict Exact Amount Validation
	if math.Abs(amount-target.AmountDue) > 0.01 {
		return fmt.Errorf("payment rejected: amount must be exactly %.2f", target.AmountDue)
	}

	// Apply Payment
	now := time.Now()
	target.IsPaid = true
	target.PaidAt = &now

	// Check if loan is fully closed
	allPaid := true
	for _, inst := range loan.Installments {
		if !inst.IsPaid {
			allPaid = false
			break
		}
	}
	if allPaid {
		loan.Status = entity.StatusClosed
	}

	return s.Repo.Update(loan)
}
