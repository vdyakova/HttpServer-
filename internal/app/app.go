package app

import (
	"HttpServer/internal/handler"
	"HttpServer/internal/middleware"
	"HttpServer/internal/repository"
	"HttpServer/internal/service"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"net/http"
	"os"
)

type App struct {
	ctx         context.Context
	pool        *pgxpool.Pool
	redisClient *redis.Client
	authService *service.AuthService
	docService  *service.DocumentService
	authHandler *handler.RegisterHandler
	docHandler  *handler.DocumentHandler
}

func NewApp(ctx context.Context) (*App, error) {

	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}
	pgDSN := os.Getenv("PG_DSN")
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	documentRepo := repository.NewDocumentRepository(pool, redisClient)
	userRepo := repository.NewUserRepository(pool)
	authService := service.NewUserService(userRepo)
	docService := service.NewDocumentService(documentRepo, authService)
	authHandler := handler.NewRegisterHandler(authService)
	docHandler := handler.NewDocumentHandler(docService, authService)

	return &App{
		ctx:         ctx,
		pool:        pool,
		redisClient: redisClient,
		authHandler: authHandler,
		docHandler:  docHandler,
	}, nil
}

func (a *App) Run() error {

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.Handle("/api/auth", middleware.WithContext(a.ctx, http.HandlerFunc(a.authHandler.Authenticate))).Methods("POST")
	r.Handle("/api/register", middleware.WithContext(a.ctx, http.HandlerFunc(a.authHandler.Register))).Methods("POST")
	r.Handle("/api/docs", middleware.WithContext(a.ctx, http.HandlerFunc(a.docHandler.GetDocuments))).Methods("GET")
	r.Handle("/api/docs/{id:[0-9]+}", middleware.WithContext(a.ctx, http.HandlerFunc(a.docHandler.GetDocumentsByID))).Methods("GET")
	r.Handle("/api/docs/{id:[0-9]+}", middleware.WithContext(a.ctx, http.HandlerFunc(a.docHandler.DeleteDoc))).Methods("DELETE")
	r.Handle("/api/docs", middleware.WithContext(a.ctx, http.HandlerFunc(a.docHandler.UploadDoc))).Methods("POST")
	r.Handle("/api/auth/{token}", middleware.WithContext(a.ctx, http.HandlerFunc(a.authHandler.DeleteToken))).Methods("DELETE")

	fmt.Println("Server started at http://localhost:8080")
	return http.ListenAndServe(":8080", r)
}
