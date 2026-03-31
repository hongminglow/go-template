package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/hongminglow/go-template/internal/pkg/httpx"
	"github.com/hongminglow/go-template/internal/user"
)

type HTTPHandler struct {
	userService *user.Service
	jwtSecret   string
}

type registerRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Gender   string `json:"gender"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewHTTPHandler(userService *user.Service, jwtSecret string) *HTTPHandler {
	return &HTTPHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdUser, err := h.userService.Create(r.Context(), user.CreateInput{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Gender:   req.Gender,
	})
	
	if err != nil {
		handleError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"data": createdUser,
	})
}

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userRecord, err := h.userService.Auth(r.Context(), req.Email, req.Password)
	if err != nil {
		handleError(w, err)
		return
	}

	token, err := httpx.GenerateJWT(h.jwtSecret, userRecord.ID, 24*time.Hour)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"data":  userRecord,
		"token": token,
	})
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case err == user.ErrUserNotFound:
		httpx.WriteError(w, http.StatusNotFound, err.Error())
	case err == user.ErrEmailAlreadyUsed || err == user.ErrUsernameAlreadyUsed:
		httpx.WriteError(w, http.StatusConflict, err.Error())
	case err == user.ErrInvalidName || err == user.ErrInvalidEmail || err == user.ErrInvalidPassword:
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	case err == user.ErrInvalidCredentials:
		httpx.WriteError(w, http.StatusUnauthorized, err.Error())
	default:
		log.Printf("unexpected auth error: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
