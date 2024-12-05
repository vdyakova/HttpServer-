package service

import (
	"HttpServer/internal/models"
	"HttpServer/internal/repository"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type DocumentService interface {
	GetDocuments(ctx context.Context, token, filterLogin, key, value string, limit int) ([]models.Document, error)
	GetDocumentById(ctx context.Context, token string, id int) (*models.Document, error)
	DeleteDoc(ctx context.Context, token string, id int) (bool, error)
	UploadDocument(ctx context.Context, doc models.Document, fileData []byte, filename string) error
}
type dockserv struct {
	docsRepository repository.DocumentRepository
	authService    AuthService
}

func NewDocumentService(repo repository.DocumentRepository, auth AuthService) DocumentService {
	return &dockserv{
		docsRepository: repo,
		authService:    auth,
	}
}

func (s *dockserv) GetDocuments(ctx context.Context, token string, filterLogin string, key string, value string, limit int) ([]models.Document, error) {

	var ownerLogin string
	if filterLogin == "" {
		var err error
		ownerLogin, err = s.authService.GetLoginFromToken(ctx, token)
		if err != nil {
			return nil, err
		}
		filterLogin = ownerLogin
	} else {
		ownerLogin = filterLogin
	}
	docs, err := s.docsRepository.FindDocuments(ctx, ownerLogin, filterLogin, key, value, limit)
	if err != nil {
		return nil, err
	}
	return docs, nil
}
func (s *dockserv) GetDocumentById(ctx context.Context, token string, id int) (*models.Document, error) {

	login, err := s.authService.GetLoginFromToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get login from token: %w", err)
	}
	doc, err := s.docsRepository.FindDocumentByID(ctx, login, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find document by ID: %w", err)
	}
	return doc, nil
}
func (s *dockserv) DeleteDoc(ctx context.Context, token string, id int) (bool, error) {
	login, err := s.authService.GetLoginFromToken(ctx, token)
	if err != nil {
		return false, fmt.Errorf("failed to get login from token: %w", err)
	}
	st, err := s.docsRepository.DeleteDoc(ctx, login, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete document: %w", err)
	}
	return st, nil
}
func (s *dockserv) UploadDocument(ctx context.Context, doc models.Document, fileData []byte, filename string) error {
	savePath := filepath.Join("uploads", filename)
	if err := os.WriteFile(savePath, fileData, 0644); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	doc.File = true

	if err := s.docsRepository.SaveDocument(ctx, doc); err != nil {
		return err
	}
	return nil
}
