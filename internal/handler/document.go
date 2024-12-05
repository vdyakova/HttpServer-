package handler

import (
	"HttpServer/internal/models"
	"HttpServer/internal/service"
	"HttpServer/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DocumentHandler struct {
	documentService service.DocumentService
	authService     service.AuthService
}

func NewDocumentHandler(docService service.DocumentService, authService service.AuthService) *DocumentHandler {
	return &DocumentHandler{

		documentService: docService,
		authService:     authService,
	}
}

func (h *DocumentHandler) GetDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var requestData struct {
		Token string `json:"token"`
		Login string `json:"login"`
		Key   string `json:"key"`
		Value string `json:"value"`
		Limit int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		utils.ErrorResponse(w, 400, "Invalid JSON", http.StatusBadRequest)
		return
	}
	docs, err := h.documentService.GetDocuments(ctx, requestData.Token, requestData.Login, requestData.Key, requestData.Value, requestData.Limit)
	fmt.Println("get запрос", requestData.Token, requestData.Login, requestData.Key, requestData.Value, requestData.Limit)
	if err != nil {
		http.Error(w, "failed to get documents", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"docs": docs,
		},
	})
}
func (h *DocumentHandler) GetDocumentsByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqToken struct {
		Token string `json:"token"`
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/docs/")
	fmt.Println(id)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.ErrorResponse(w, 400, "Неверный формат ID", http.StatusBadRequest)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&reqToken); err != nil {
		fmt.Println(err)
		utils.ErrorResponse(w, 400, "Invalid JSON", http.StatusBadRequest)
		return
	}

	doc, err := h.documentService.GetDocumentById(ctx, reqToken.Token, idInt)
	if err != nil {
		http.Error(w, "failed to get documents", http.StatusInternalServerError)
		return
	}
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"docs": doc,
		},
	})
}
func (h *DocumentHandler) DeleteDoc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqToken struct {
		Token string `json:"token"`
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/docs/")
	idInt, err := strconv.Atoi(id)
	if err := json.NewDecoder(r.Body).Decode(&reqToken); err != nil {
		fmt.Println(err)
		utils.ErrorResponse(w, 400, "Invalid JSON", http.StatusBadRequest)
		return
	}
	st, err := h.documentService.DeleteDoc(ctx, reqToken.Token, idInt)
	if err != nil {
		http.Error(w, "failed to get documents", http.StatusInternalServerError)
		return
	}
	fmt.Println(st)
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			id: st,
		},
	})
}

func (h *DocumentHandler) UploadDoc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	docStr := r.FormValue("document")
	var doc models.Document
	if err := json.Unmarshal([]byte(docStr), &doc); err != nil {
		http.Error(w, "Invalid document JSON", http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileData := make([]byte, handler.Size)
	_, err = file.Read(fileData)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	err = h.documentService.UploadDocument(ctx, doc, fileData, handler.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload document: %s", err), http.StatusInternalServerError)
		return
	}
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"file": handler.Filename,
			"json": doc,
		},
	})
}

func saveFile(file io.Reader, fileName string) (string, error) {
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", errors.New("failed to create upload directory")
	}

	filePath := filepath.Join(uploadDir, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return "", errors.New("failed to create file")
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		return "", errors.New("failed to save file")
	}

	return filePath, nil
}
