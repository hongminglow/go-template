package user

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hongminglow/go-template/internal/pkg/httpx"
)

type HTTPHandler struct {
	service   *Service
}

type updateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{
		service:   service,
	}
}



func (h *HTTPHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.GetUserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userRecord, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		handleUserError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"data": userRecord,
	})
}

func (h *HTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := parseQueryInt32(r, "limit")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "query parameter 'limit' must be a number")
		return
	}

	offset, err := parseQueryInt32(r, "offset")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "query parameter 'offset' must be a number")
		return
	}

	users, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		handleUserError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"data": users,
	})
}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	userRecord, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		handleUserError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"data": userRecord,
	})
}

func (h *HTTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	var req updateUserRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedUser, err := h.service.Update(r.Context(), userID, UpdateInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		handleUserError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"data": updatedUser,
	})
}

func (h *HTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), userID); err != nil {
		handleUserError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	rawUserID := chi.URLParam(r, "userID")
	userID, err := strconv.ParseInt(rawUserID, 10, 64)
	if err != nil || userID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "path parameter 'userID' must be a positive integer")
		return 0, false
	}

	return userID, true
}

func parseQueryInt32(r *http.Request, key string) (int32, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(value), nil
}

func handleUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUserNotFound):
		httpx.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrEmailAlreadyUsed), errors.Is(err, ErrUsernameAlreadyUsed):
		httpx.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidName), errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrInvalidPassword):
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, err.Error())
	default:
		log.Printf("unexpected user module error: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
