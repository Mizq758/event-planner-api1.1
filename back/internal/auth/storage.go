package auth

import (
    "context"
    "errors"
    "search-job/internal/models"
    "search-job/pkg/postgres"
    "github.com/jackc/pgx/v5"
)

type Repo struct {
    db *postgres.DB
}

func NewRepo(db *postgres.DB) *Repo {
    return &Repo{db: db}
}

// CreateUser создает нового пользователя
func (r *Repo) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
    var user models.User
    query := `
        INSERT INTO users (email, password_hash, created_at)
        VALUES ($1, $2, NOW())
        RETURNING id, email, password_hash, created_at
    `

    err := r.db.QueryRow(ctx, query, email, passwordHash).Scan(
        &user.ID,
        &user.Email,
        &user.PasswordHash,
        &user.CreatedAt,
    )

    if err != nil {
        return nil, err
    }

    return &user, nil
}

// GetUserByEmail получает пользователя по email
func (r *Repo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    query := `
        SELECT id, email, password_hash, created_at
        FROM users
        WHERE email = $1
    `

    err := r.db.QueryRow(ctx, query, email).Scan(
        &user.ID,
        &user.Email,
        &user.PasswordHash,
        &user.CreatedAt,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("user not found")
        }
        return nil, err
    }

    return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *Repo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
    var user models.User
    query := `
        SELECT id, email, password_hash, created_at
        FROM users
        WHERE id = $1
    `

    err := r.db.QueryRow(ctx, query, id).Scan(
        &user.ID,
        &user.Email,
        &user.PasswordHash,
        &user.CreatedAt,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("user not found")
        }
        return nil, err
    }

    return &user, nil
}