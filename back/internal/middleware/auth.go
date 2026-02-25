package middleware

import (
    "net/http"
    "search-job/internal/service"  // меняем auth на service
    "search-job/pkg/jwt"
    "strings"

    "github.com/labstack/echo/v4"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(svc *service.Service) echo.MiddlewareFunc {  // меняем тип
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            authHeader := c.Request().Header.Get("Authorization")
            if authHeader == "" {
                return c.JSON(http.StatusUnauthorized, map[string]interface{}{
                    "error":  "authorization header required",
                    "status": "error",
                })
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
                return c.JSON(http.StatusUnauthorized, map[string]interface{}{
                    "error":  "invalid authorization header format. Use: Bearer <token>",
                    "status": "error",
                })
            }

            token := parts[1]

            userID, err := jwt.ValidateToken(token)
            if err != nil {
                return c.JSON(http.StatusUnauthorized, map[string]interface{}{
                    "error":  "invalid or expired token",
                    "status": "error",
                })
            }

            // Проверяем существование пользователя через сервис
            user, err := svc.GetUserByID(c.Request().Context(), userID)
            if err != nil || user == nil {
                return c.JSON(http.StatusUnauthorized, map[string]interface{}{
                    "error":  "user not found",
                    "status": "error",
                })
            }

            c.Set("user_id", userID)
            c.Set("user", user)

            return next(c)
        }
    }
}

