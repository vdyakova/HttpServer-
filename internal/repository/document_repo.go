package repository

import (
	"HttpServer/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"time"
)

type DocumentRepository interface {
	FindDocuments(ctx context.Context, ownerLogin, filterLogin, key, value string, limit int) ([]models.Document, error)
	FindDocumentByID(ctx context.Context, ownerLogin string, ID int) (*models.Document, error)
	DeleteDoc(ctx context.Context, login string, id int) (bool, error)
	SaveDocument(ctx context.Context, doc models.Document) error
}

type repo struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewDocumentRepository(db *pgxpool.Pool, redis *redis.Client) DocumentRepository {
	return &repo{db: db, redis: redis}
}
func (r *repo) FindDocuments(ctx context.Context, ownerLogin, filterLogin, key, value string, limit int) ([]models.Document, error) {
	cacheKey := fmt.Sprintf("documents:%s:%s:%d", ownerLogin, filterLogin, limit)

	cachedDocs, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var docs []models.Document
		if err := json.Unmarshal([]byte(cachedDocs), &docs); err == nil {
			return docs, nil
		}
	}

	var docs []models.Document
	query := `SELECT id, name, mime, file, "public", owner_login, created, grants 
          FROM documents5 
          WHERE (owner_login = $1 OR $2 = ANY(grants)) 
          ORDER BY name, created 
          LIMIT $3`

	params := []interface{}{ownerLogin, filterLogin, limit}
	rows, err := r.db.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var doc models.Document
		var doc2 string
		if err := rows.Scan(&doc.ID, &doc.Name, &doc.Mime, &doc.File, &doc.Public, &doc2, &doc.Created, &doc.Grant); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	data, err := json.Marshal(docs)
	if err == nil {
		r.redis.Set(ctx, cacheKey, data, time.Minute*10)
	}

	return docs, nil
}

func (r *repo) FindDocumentByID(ctx context.Context, ownerLogin string, ID int) (*models.Document, error) {
	cacheKey := fmt.Sprintf("document:%d:%s", ID, ownerLogin)

	cachedDoc, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var doc models.Document
		if err := json.Unmarshal([]byte(cachedDoc), &doc); err == nil {
			return &doc, nil
		}
	}

	var doc models.Document
	var doc2 string
	query := `
        SELECT id, name, mime, file, "public", owner_login, created, grants
        FROM documents5
        WHERE id = $1 AND (owner_login = $2 OR $2 = ANY(grants));
    `
	err = r.db.QueryRow(ctx, query, ID, ownerLogin).Scan(
		&doc.ID, &doc.Name, &doc.Mime, &doc.File, &doc.Public, &doc2, &doc.Created, &doc.Grant,
	)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(doc)
	if err == nil {
		r.redis.Set(ctx, cacheKey, data, time.Minute*10)
	}

	return &doc, nil
}

func (r *repo) DeleteDoc(ctx context.Context, login string, id int) (bool, error) {
	query := `delete from documents5 where (owner_login = $1 OR $1 = ANY(grants)) AND id=$2`
	res, err := r.db.Exec(ctx, query, login, id)
	rowsAffected := res.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, nil
	}

	r.redis.Del(ctx, fmt.Sprintf("document:%d:%s", id, login))

	return true, nil
}

func (r *repo) SaveDocument(ctx context.Context, doc models.Document) error {
	query := `
		INSERT INTO documents5 (name, mime, file, public, created, grant)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, doc.Name, doc.Mime, doc.File, doc.Public, time.Now(), pq.Array(doc.Grant))
	if err != nil {
		return fmt.Errorf("failed to save document in database: %w", err)
	}

	r.redis.FlushDB(ctx)

	return nil
}
