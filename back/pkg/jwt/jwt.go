package jwt

import (
    "errors"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "os"
)

var jwtSecret = []byte(getSecretKey())

func getSecretKey() string {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        // Только для разработки! В проде всегда через env
        secret = "your-secret-key-change-in-production"
    }
    return secret
}

type Claims struct {
    UserID int64 `json:"user_id"`
    jwt.RegisteredClaims
}

// GenerateToken создает JWT токен для пользователя
func GenerateToken(userID int64) (string, error) {
    claims := Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// ValidateToken проверяет токен и возвращает user_id
func ValidateToken(tokenString string) (int64, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return jwtSecret, nil
    })

    if err != nil {
        return 0, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims.UserID, nil
    }

    return 0, errors.New("invalid token")
}