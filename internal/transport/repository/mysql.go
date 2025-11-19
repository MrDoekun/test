package repository

import (
	"amartha-test/internal/entity"
	"amartha-test/internal/usecase"
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

// Ensure interface compliance
var _ usecase.LoanRepository = (*MySQLLoanRepository)(nil)

type MySQLLoanRepository struct {
	db *sql.DB
}

func NewMySQLLoanRepository(connStr string) (*MySQLLoanRepository, error) {
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &MySQLLoanRepository{db: db}, nil
}

func (r *MySQLLoanRepository) Save(loan *entity.Loan) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Loan Header
	query := `INSERT INTO loans (principal_amount, interest_rate, total_payable, weeks_duration, start_date, status) 
	          VALUES (?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(query, loan.Principal, loan.InterestRate, loan.TotalPayable, loan.WeeksDuration, loan.StartDate, loan.Status)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	loan.ID = int(id)

	// 2. Insert Installments (Schedule)
	instQuery := `INSERT INTO installments (loan_id, week_number, due_date, amount_due, is_paid) VALUES (?, ?, ?, ?, ?)`
	stmt, err := tx.Prepare(instQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, inst := range loan.Installments {
		_, err = stmt.Exec(loan.ID, inst.WeekNumber, inst.DueDate, inst.AmountDue, inst.IsPaid)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MySQLLoanRepository) FindByID(id int) (*entity.Loan, error) {
	loan := &entity.Loan{}

	// 1. Get Loan Header
	err := r.db.QueryRow("SELECT id, principal_amount, interest_rate, total_payable, weeks_duration, start_date, status FROM loans WHERE id = ?", id).
		Scan(&loan.ID, &loan.Principal, &loan.InterestRate, &loan.TotalPayable, &loan.WeeksDuration, &loan.StartDate, &loan.Status)

	if err == sql.ErrNoRows {
		return nil, errors.New("loan not found")
	} else if err != nil {
		return nil, err
	}

	// 2. Get Installments
	rows, err := r.db.Query("SELECT id, loan_id, week_number, due_date, amount_due, is_paid, paid_at FROM installments WHERE loan_id = ? ORDER BY week_number ASC", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		inst := &entity.Installment{}
		var paidAt sql.NullTime // Handle nullable column

		err := rows.Scan(&inst.ID, &inst.LoanID, &inst.WeekNumber, &inst.DueDate, &inst.AmountDue, &inst.IsPaid, &paidAt)
		if err != nil {
			return nil, err
		}

		if paidAt.Valid {
			t := paidAt.Time
			inst.PaidAt = &t
		}

		loan.Installments = append(loan.Installments, inst)
	}

	return loan, nil
}

func (r *MySQLLoanRepository) Update(loan *entity.Loan) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update Loan Status
	_, err = tx.Exec("UPDATE loans SET status = ? WHERE id = ?", loan.Status, loan.ID)
	if err != nil {
		return err
	}

	// 2. Update Installments
	// Optimization: In a real high-throughput system, we'd only update changed rows.
	// Here, for simplicity and consistency, we update the state of payments.
	stmt, err := tx.Prepare("UPDATE installments SET is_paid = ?, paid_at = ? WHERE loan_id = ? AND week_number = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, inst := range loan.Installments {
		_, err = stmt.Exec(inst.IsPaid, inst.PaidAt, loan.ID, inst.WeekNumber)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
