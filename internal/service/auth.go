package service

import (
	"HttpServer/internal/repository"
	"HttpServer/internal/utils"
	"context"
	"errors"
	"fmt"
	"log"
)

type AuthService interface {
	RegisterUser(ctx context.Context, login, password string) error
	Authenticate(ctx context.Context, login, password string) (string, error)
	GetLoginFromToken(ctx context.Context, token string) (string, error)
	DeleteToken(ctx context.Context, token string) (bool, error)
}

type authstvc struct {
	userRepo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) AuthService {
	return &authstvc{userRepo: repo}
}

func (s *authstvc) RegisterUser(ctx context.Context, login, password string) error {
	exists, err := s.userRepo.UserExists(ctx, login)
	if err != nil {
		fmt.Println("err reg user", err)
		return err
	}
	if exists {
		return errors.New("user already exists")
	}
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	err = s.userRepo.SaveUser(ctx, login, hashedPassword)
	if err != nil {
		log.Fatalf("Failed to save user: %v", err)
		return err
	}
	log.Println("User saved successfully!")
	return nil
}

func hashPassword(password string) (string, error) {
	return password, nil
}

func (s *authstvc) Authenticate(ctx context.Context, login string, password string) (string, error) {
	token, err := utils.GenerateToken(login)
	fmt.Println("login from service", login)
	if err != nil {
		return "", fmt.Errorf("failed to generate token for login %s: %w", login, err)
	}
	if err := s.userRepo.SaveToken(ctx, login, token); err != nil {
		fmt.Printf("failed to save token for login %s: %v\n", login, err)
		return "", fmt.Errorf("failed to save token: %w", err)
	}
	return token, nil
}
func (s *authstvc) GetLoginFromToken(ctx context.Context, token string) (string, error) {
	login, err := s.userRepo.GetLoginFromToken(ctx, token)
	if err != nil {
		fmt.Println(err)
	}
	return login, nil
}
func (h *authstvc) DeleteToken(ctx context.Context, token string) (bool, error) {
	st, err := h.userRepo.DeleteToken(ctx, token)
	if err != nil {
		fmt.Println(err)
	}
	return st, nil
}
