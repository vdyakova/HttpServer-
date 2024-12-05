package handler

import (
	"HttpServer/internal/service"
	"HttpServer/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type RegisterHandler struct {
	service service.AuthService
}

func NewRegisterHandler(svc service.AuthService) *RegisterHandler {
	return &RegisterHandler{

		service: svc,
	}
}

func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, 405, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var requestData struct {
		Token string `json:"token"`
		Login string `json:"login"`
		Pswd  string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		utils.ErrorResponse(w, 400, "Invalid JSON", http.StatusBadRequest)
		return
	}

	adminToken := os.Getenv("TOKEN_ADMIN")
	if requestData.Token != adminToken {
		utils.ErrorResponse(w, 403, "Invalid admin token", http.StatusForbidden)
		return
	}
	if len(requestData.Login) < 8 || !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(requestData.Login) {
		utils.ErrorResponse(w, 400, "Invalid login format", http.StatusBadRequest)
		return
	}

	if err := validatePassword(requestData.Pswd); err != nil {

		utils.ErrorResponse(w, 400, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.RegisterUser(ctx, requestData.Login, requestData.Pswd); err != nil {
		utils.ErrorResponse(w, 500, "Failed to register user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"response": map[string]string{
			"login": requestData.Login,
		},
	}
	utils.SuccessResponse(w, response, "User registered successfully")
}

func validatePassword(password string) error {
	if len(password) < 8 {
		fmt.Println(len(password))
		return errors.New("Password must be at least 8 characters long")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("Password must contain at least one uppercase letter")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("Password must contain at least one lowercase letter")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("Password must contain at least one digit")
	}
	if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		return errors.New("Password must contain at least one special character")
	}
	return nil
}

func (h *RegisterHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		return
	}
	var credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		return
	}
	fmt.Println("from handler login", credentials.Login)
	token, err := h.service.Authenticate(ctx, credentials.Login, credentials.Password)
	if err != nil {
		return
	}
	response := map[string]interface{}{
		"response": map[string]string{
			"token": token,
		},
	}
	utils.SuccessResponse(w, response, "Authentication successful")
}

func (h *RegisterHandler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("Request URL: %s, Method: %s\n", r.URL.Path, r.Method)
	token := strings.TrimPrefix(r.URL.Path, "/api/auth/")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}
	fmt.Println("Received token:", token)
	st, err := h.service.DeleteToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to delete token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"response": map[string]bool{
			token: st,
		},
	}
	utils.SuccessResponse(w, response, "Token deleted successfully")
}
