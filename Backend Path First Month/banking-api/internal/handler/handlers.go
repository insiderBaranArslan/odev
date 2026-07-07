package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"banking-api/internal/domain"
	"banking-api/internal/middleware"
	"banking-api/internal/service"
	"banking-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	users *service.UserService
}

func NewAuthHandler(users *service.UserService) *AuthHandler {
	return &AuthHandler{users: users}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.Register(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	tokens, err := h.users.IssueTokens(r.Context(), user)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"user": user, "tokens": tokens})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.Authenticate(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	tokens, err := h.users.IssueTokens(r.Context(), user)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tokens, err := h.users.RefreshTokens(r.Context(), body.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, tokens)
}

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.ListUsers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.users.GetUser(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.UpdateUser(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.users.DeleteUser(r.Context(), id); err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type TransactionHandler struct {
	transactions *service.TransactionService
}

func NewTransactionHandler(transactions *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactions: transactions}
}

func (h *TransactionHandler) Credit(w http.ResponseWriter, r *http.Request) {
	var req domain.CreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tx, err := h.transactions.Credit(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) {
	var req domain.DebitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tx, err := h.transactions.Debit(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req domain.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tx, err := h.transactions.Transfer(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) History(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if queryID := r.URL.Query().Get("user_id"); queryID != "" {
		parsed, err := parseID(queryID)
		if err == nil {
			userID = parsed
		}
	}
	txs, err := h.transactions.GetHistory(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, txs)
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	tx, err := h.transactions.GetTransaction(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) Stats(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.transactions.Stats())
}

type BalanceHandler struct {
	balances *service.BalanceService
}

func NewBalanceHandler(balances *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{balances: balances}
}

func (h *BalanceHandler) Current(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if queryID := r.URL.Query().Get("user_id"); queryID != "" {
		parsed, err := parseID(queryID)
		if err == nil {
			userID = parsed
		}
	}
	balance, err := h.balances.GetCurrent(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, balance)
}

func (h *BalanceHandler) Historical(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r.URL.Query().Get("user_id"))
	if err != nil {
		userID, _ = middleware.UserIDFromContext(r.Context())
	}
	snapshots, err := h.balances.GetHistorical(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, snapshots)
}

func (h *BalanceHandler) AtTime(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r.URL.Query().Get("user_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "user_id is required")
		return
	}
	atRaw := r.URL.Query().Get("at")
	at, err := time.Parse(time.RFC3339, atRaw)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid at timestamp, use RFC3339")
		return
	}
	snapshot, err := h.balances.GetAtTime(r.Context(), userID, at)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, snapshot)
}

func Health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
