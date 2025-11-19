package entity

import "time"

type LoanStatus string

const (
	StatusActive     LoanStatus = "ACTIVE"
	StatusClosed     LoanStatus = "CLOSED"
	StatusDelinquent LoanStatus = "DELINQUENT"
)

type Loan struct {
	ID            int            `json:"id"`
	Principal     float64        `json:"principal"`
	InterestRate  float64        `json:"interest_rate"`
	TotalPayable  float64        `json:"total_payable"`
	WeeksDuration int            `json:"weeks_duration"`
	StartDate     time.Time      `json:"start_date"`
	Status        LoanStatus     `json:"status"`
	Installments  []*Installment `json:"installments"`
}

type Installment struct {
	ID         int        `json:"id"`
	LoanID     int        `json:"loan_id"`
	WeekNumber int        `json:"week_number"`
	DueDate    time.Time  `json:"due_date"`
	AmountDue  float64    `json:"amount_due"`
	IsPaid     bool       `json:"is_paid"`
	PaidAt     *time.Time `json:"paid_at,omitempty"` // Pointer to handle null
}
