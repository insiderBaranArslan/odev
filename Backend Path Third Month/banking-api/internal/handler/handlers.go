package handler

import (
	"encoding/json"
	"io"
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
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.users.Register(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	tokens, err := h.users.IssueTokens(user)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"user": user, "tokens": tokens})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.users.Authenticate(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	tokens, err := h.users.IssueTokens(user)
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
	if err := decodeJSON(r, &body); err != nil || body.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}
	tokens, err := h.users.RefreshTokens(body.RefreshToken)
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
	users, err := h.users.List(r.Context())
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
	user, err := h.users.GetByID(r.Context(), id)
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
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.users.Update(r.Context(), id, req)
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
	if err := h.users.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type TransactionHandler struct {
	txs *service.TransactionService
}

func NewTransactionHandler(txs *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{txs: txs}
}

func (h *TransactionHandler) Credit(w http.ResponseWriter, r *http.Request) {
	var req domain.CreditRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	tx, err := h.txs.Credit(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) {
	var req domain.DebitRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	role, _ := middleware.RoleFromContext(r.Context())
	if role != domain.RoleAdmin && req.UserID != userID {
		response.Error(w, http.StatusForbidden, "cannot debit another user")
		return
	}
	tx, err := h.txs.Debit(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req domain.TransferRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	role, _ := middleware.RoleFromContext(r.Context())
	if role != domain.RoleAdmin && req.FromUserID != userID {
		response.Error(w, http.StatusForbidden, "cannot transfer from another account")
		return
	}
	tx, err := h.txs.Transfer(r.Context(), req)
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
	role, _ := middleware.RoleFromContext(r.Context())
	if q := r.URL.Query().Get("user_id"); q != "" {
		parsed, err := parseID(q)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		if role != domain.RoleAdmin && parsed != userID {
			response.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		userID = parsed
	}
	txs, err := h.txs.History(r.Context(), userID)
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
	tx, err := h.txs.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, tx)
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
	role, _ := middleware.RoleFromContext(r.Context())
	if q := r.URL.Query().Get("user_id"); q != "" {
		parsed, err := parseID(q)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		if role != domain.RoleAdmin && parsed != userID {
			response.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		userID = parsed
	}
	balance, err := h.balances.GetCurrent(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, balance.Snapshot())
}

func (h *BalanceHandler) Historical(w http.ResponseWriter, r *http.Request) {
	userID, err := resolveUserID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	snaps, err := h.balances.GetHistorical(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, snaps)
}

func (h *BalanceHandler) AtTime(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r.URL.Query().Get("user_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "user_id is required")
		return
	}
	at, err := time.Parse(time.RFC3339, r.URL.Query().Get("at"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid at timestamp, use RFC3339")
		return
	}
	snap, err := h.balances.GetAtTime(r.Context(), userID, at)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, snap)
}

func Health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}

func resolveUserID(r *http.Request) (int64, error) {
	if q := r.URL.Query().Get("user_id"); q != "" {
		return parseID(q)
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return 0, strconv.ErrSyntax
	}
	return userID, nil
}
