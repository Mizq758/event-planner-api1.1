
package models

import "time"

type Events struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Location    string    `json:"location" db:"location"`
	StartAt     time.Time `json:"startAt" db:"start_at"`
	EndAt       time.Time `json:"endAt" db:"end_at"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	DeletedAt   time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

// UpdateEventRequest для частичного обновления (PATCH)
type UpdateEventRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Location    *string    `json:"location,omitempty"`
	StartAt     *time.Time `json:"startAt,omitempty"`
	EndAt       *time.Time `json:"endAt,omitempty"`
}
// User модель пользователя
type User struct {
    ID           int64     `json:"id" db:"id"`
    Email        string    `json:"email" db:"email"`
    PasswordHash string    `json:"-" db:"password_hash"` // "-" скрывает в JSON
    CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

// RegisterRequest для регистрации
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest для входа
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// AuthResponse ответ с токеном
type AuthResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

// Participant модель участника
type Participant struct {
    ID        int64     `json:"id" db:"id"`
    EventID   int64     `json:"event_id" db:"event_id"`
    Name      string    `json:"name" db:"name"`
    Email     *string   `json:"email,omitempty" db:"email"`
    CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// ParticipantCreate для создания участника
type ParticipantCreate struct {
    Name  string  `json:"name" validate:"required"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

// EventFilter для фильтрации списка событий
type EventFilter struct {
    From   *time.Time `query:"from"`
    To     *time.Time `query:"to"`
    Search *string    `query:"search"`
    Page   int        `query:"page" default:"1"`
    Limit  int        `query:"limit" default:"10"`
    Sort   string     `query:"sort" default:"created_at"` // start_at или created_at
    Order  string     `query:"order" default:"desc"`      // asc или desc
}