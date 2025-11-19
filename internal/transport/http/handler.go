package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"amartha-test/internal/usecase"
)

type LoanHandler struct {
	service *usecase.BillingService
}

func NewLoanHandler(service *usecase.BillingService) *LoanHandler {
	return &LoanHandler{service: service}
}

type CreateLoanRequest struct {
	Principal float64 `json:"principal"`
	Weeks     int     `json:"weeks"`
}

type PaymentRequest struct {
	Amount float64 `json:"amount"`
}

func (h *LoanHandler) CreateLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	loan, err := h.service.CreateLoan(req.Principal, 0.10, req.Weeks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loan)
}

func (h *LoanHandler) GetLoanDetails(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	isDelinquent, err := h.service.CheckDelinquency(id)
	if err != nil {
		http.Error(w, "Loan not found", http.StatusNotFound)
		return
	}

	outstanding, _ := h.service.GetOutstanding(id)
	loan, _ := h.service.Repo.FindByID(id)

	response := map[string]interface{}{
		"loan_id":       id,
		"status":        loan.Status,
		"outstanding":   outstanding,
		"is_delinquent": isDelinquent,
		"installments":  loan.Installments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LoanHandler) MakePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.MakePayment(id, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Payment Accepted"})
}
