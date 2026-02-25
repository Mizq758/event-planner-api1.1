package service

import (
    "context"
    "errors"
    "net/http"
    "search-job/internal/models"
    "golang.org/x/crypto/bcrypt"
    "search-job/pkg/jwt"

    "github.com/labstack/echo/v4"
)

// Register - регистрация нового пользователя
func (s *Service) Register(c echo.Context) error {
    var req models.RegisterRequest
    
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid request format"))
    }

    if req.Email == "" {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("email is required"))
    }
    if req.Password == "" || len(req.Password) < 6 {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("password must be at least 6 characters"))
    }

    response, err := s.registerUser(c.Request().Context(), req)
    if err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse(err.Error()))
    }

    return c.JSON(http.StatusCreated, s.SuccessResponse(response))
}

// Login - вход пользователя
func (s *Service) Login(c echo.Context) error {
    var req models.LoginRequest
    
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid request format"))
    }

    if req.Email == "" {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("email is required"))
    }
    if req.Password == "" {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("password is required"))
    }

    response, err := s.loginUser(c.Request().Context(), req)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse(err.Error()))
    }

    return c.JSON(http.StatusOK, s.SuccessResponse(response))
}

// GetProfile - получение профиля (для проверки токена)
func (s *Service) GetProfile(c echo.Context) error {
    // Получаем user_id из контекста (устанавливается middleware)
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }
    
    userID, ok := userIDVal.(int64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }
    
    user, err := s.getUserByID(c.Request().Context(), userID)
    if err != nil {
        return c.JSON(http.StatusNotFound, s.ErrorResponse("user not found"))
    }

    user.PasswordHash = ""
    return c.JSON(http.StatusOK, s.SuccessResponse(user))
}

// Внутренние методы (бизнес-логика)

func (s *Service) registerUser(ctx context.Context, req models.RegisterRequest) (*models.AuthResponse, error) {
    // Проверяем существование пользователя
    existingUser, err := s.authRepo.GetUserByEmail(ctx, req.Email)
    if err == nil && existingUser != nil {
        return nil, errors.New("user with this email already exists")
    }

    // Хешируем пароль
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.New("failed to hash password")
    }

    // Создаём пользователя
    user, err := s.authRepo.CreateUser(ctx, req.Email, string(hashedPassword))
    if err != nil {
        return nil, err
    }

    // Генерируем токен
    token, err := jwt.GenerateToken(user.ID)
    if err != nil {
        return nil, errors.New("failed to generate token")
    }

    user.PasswordHash = ""
    return &models.AuthResponse{
        Token: token,
        User:  *user,
    }, nil
}

func (s *Service) loginUser(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
    // Ищем пользователя
    user, err := s.authRepo.GetUserByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.New("invalid email or password")
    }

    // Проверяем пароль
    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
    if err != nil {
        return nil, errors.New("invalid email or password")
    }

    // Генерируем токен
    token, err := jwt.GenerateToken(user.ID)
    if err != nil {
        return nil, errors.New("failed to generate token")
    }

    user.PasswordHash = ""
    return &models.AuthResponse{
        Token: token,
        User:  *user,
    }, nil
}

func (s *Service) getUserByID(ctx context.Context, id int64) (*models.User, error) {
    return s.authRepo.GetUserByID(ctx, id)
}