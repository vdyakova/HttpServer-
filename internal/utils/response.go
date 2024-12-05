package utils

import (
	"HttpServer/internal/models"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

var jwtSecret = []byte("1234")

func SuccessResponse(w http.ResponseWriter, data interface{}, responseMessage string) {
	apiResponse := models.APIResponse{
		Response: &models.ActionConfirm{
			Message: responseMessage,
		},
		Data: data,
	}

	writeJSON(w, http.StatusOK, apiResponse)
}
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}
}

func ErrorResponse(w http.ResponseWriter, code int, message string, httpStatus int) {
	apiResponse := models.APIResponse{
		Error: &models.APIError{
			Code: code,
			Text: message,
		},
	}

	writeJSON(w, httpStatus, apiResponse)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func GenerateToken(login string) (string, error) {
	claims := &jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}
