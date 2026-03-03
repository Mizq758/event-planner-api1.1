package service

import (
    "context"
    "errors"
    "search-job/internal/auth"
    "search-job/internal/events"
    "search-job/internal/models"
    "search-job/internal/participants"
    "search-job/internal/pkg/external"  // ДОБАВЛЕНО
    "github.com/labstack/echo/v4"
)

const (
    InvalidParams       = "invalid params"
    InternalServerError = "internal error"
    NotFound            = "not found"
    Conflict            = "conflict"
)

type Service struct {
    logger          echo.Logger
    eventsRepo      *events.Repo
    authRepo        *auth.Repo
    participantRepo *participants.Repo
    externalClient  *external.Client  // ДОБАВЛЕНО
}

func NewService(
    logger echo.Logger,
    eventsRepo *events.Repo,
    authRepo *auth.Repo,
    participantRepo *participants.Repo,
    externalClient *external.Client,  // ДОБАВЛЕНО
) *Service {
    return &Service{
        logger:          logger,
        eventsRepo:      eventsRepo,
        authRepo:        authRepo,
        participantRepo: participantRepo,
        externalClient:  externalClient,  // ДОБАВЛЕНО
    }
}

// Response структура для ответов
type Response struct {
    Items        interface{} `json:"items,omitempty"`
    Total        int         `json:"total,omitempty"`
    Page         int         `json:"page,omitempty"`
    Limit        int         `json:"limit,omitempty"`
    TotalPages   int         `json:"total_pages,omitempty"`
    ErrorMessage string      `json:"error,omitempty"`
    Status       string      `json:"status,omitempty"`
}

// SuccessResponse для успешных ответов
func (s *Service) SuccessResponse(object interface{}) *Response {
    return &Response{
        Items:  object,
        Status: "success",
    }
}

// ErrorResponse для ошибок
func (s *Service) ErrorResponse(message string) *Response {
    return &Response{
        ErrorMessage: message,
        Status:      "error",
    }
}

// ListResponse для списков
func (s *Service) ListResponse(items interface{}, total, page, limit int) *Response {
    totalPages := (total + limit - 1) / limit
    if totalPages < 1 {
        totalPages = 1
    }
    return &Response{
        Items:      items,
        Total:      total,
        Page:       page,
        Limit:      limit,
        TotalPages: totalPages,
        Status:     "success",
    }
}

// NewError создает ошибку (для обратной совместимости с events.go)
func (s *Service) NewError(err string) (int, *Response) {
    statusCode := 400
    switch err {
    case InvalidParams:
        statusCode = 400
    case NotFound:
        statusCode = 404
    case Conflict:
        statusCode = 409
    case InternalServerError:
        statusCode = 500
    }
    return statusCode, &Response{
        ErrorMessage: err,
        Status:      "error",
    }
}

// GetUserByID - публичный метод для получения пользователя по ID
func (s *Service) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
    if s.authRepo == nil {
        return nil, errors.New("auth repo not initialized")
    }
    return s.authRepo.GetUserByID(ctx, id)
}