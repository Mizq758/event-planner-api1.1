package events

import (
	"context"
	"errors"
	"strconv"
	"time"

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

// 1. CREATE - Создание события
func (r *Repo) RCreateEvents(ctx context.Context, events *models.Events) error {
	query := `INSERT INTO events (user_id, title, description, location, start_at, end_at, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) 
	          RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		events.UserID,
		events.Title,
		events.Description,
		events.Location,
		events.StartAt,
		events.EndAt,
	).Scan(&events.ID, &events.CreatedAt, &events.UpdatedAt)
	
	if err != nil {
		return err
	}
	return nil
}

// 2. GET BY ID - Получение события по ID (ID теперь int64)
func (r *Repo) RGetEventsById(ctx context.Context, id int64) (*models.Events, error) {
	var events models.Events
	query := `SELECT id, user_id, title, description, location, start_at, end_at, created_at, updated_at 
	          FROM events WHERE id = $1`
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&events.ID, 
		&events.UserID, 
		&events.Title, 
		&events.Description,
		&events.Location, 
		&events.StartAt, 
		&events.EndAt, 
		&events.CreatedAt,
		&events.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}
	return &events, nil
}

// 3. UPDATE - Обновление события
func (r *Repo) RUpdateEvents(ctx context.Context, events *models.Events) error {
	query := `
		UPDATE events 
		SET title = $1, description = $2, location = $3, 
			start_at = $4, end_at = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		events.Title,
		events.Description,
		events.Location,
		events.StartAt,
		events.EndAt,
		events.ID,
	).Scan(&events.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("event not found")
		}
		return err
	}
	return nil
}

// 4. DELETE - Удаление события (ID теперь int64)
func (r *Repo) RDeleteEvents(ctx context.Context, id int64) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("event not found")
	}
	return nil
}

// 5. LIST - Список событий с фильтрацией и пагинацией (ОБНОВЛЕНО)
func (r *Repo) RListEvents(ctx context.Context, userID int64, filter models.EventFilter) ([]models.Events, int, error) {
	offset := (filter.Page - 1) * filter.Limit

	// Базовый запрос
	baseQuery := `FROM events WHERE user_id = $1`
	args := []interface{}{userID}
	argPos := 2

	// Добавляем фильтры
	if filter.From != nil {
		baseQuery += ` AND start_at >= $` + strconv.Itoa(argPos)
		args = append(args, *filter.From)
		argPos++
	}
	if filter.To != nil {
		baseQuery += ` AND start_at <= $` + strconv.Itoa(argPos)
		args = append(args, *filter.To)
		argPos++
	}
	if filter.Search != nil && *filter.Search != "" {
		baseQuery += ` AND (title ILIKE $` + strconv.Itoa(argPos) + ` OR location ILIKE $` + strconv.Itoa(argPos) + `)`
		args = append(args, "%"+*filter.Search+"%")
		argPos++
	}

	// Получаем общее количество
	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := r.db.QueryRow(ctx, countQuery, args[:argPos-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Сортировка
	sortField := "created_at"
	if filter.Sort == "start_at" {
		sortField = "start_at"
	}
	
	orderDir := "DESC"
	if filter.Order == "asc" {
		orderDir = "ASC"
	}

	// Получаем события
	query := `
		SELECT id, user_id, title, description, location, start_at, end_at, created_at, updated_at
	` + baseQuery + `
		ORDER BY ` + sortField + ` ` + orderDir + `
		LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
	
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []models.Events
	for rows.Next() {
		var e models.Events
		err := rows.Scan(
			&e.ID,
			&e.UserID,
			&e.Title,
			&e.Description,
			&e.Location,
			&e.StartAt,
			&e.EndAt,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}

	return events, total, nil
}

// Вспомогательный метод для проверки дубликатов (userID теперь int64)
func (r *Repo) RCheckDuplicate(ctx context.Context, userID int64, title string, startAt time.Time) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM events WHERE user_id = $1 AND title = $2 AND start_at = $3)`
	err := r.db.QueryRow(ctx, query, userID, title, startAt).Scan(&exists)
	return exists, err
}

// Получение событий по пользователю (userID теперь int64)
func (r *Repo) RGetEventsByUser(ctx context.Context, userID int64) ([]models.Events, error) {
	query := `
		SELECT id, user_id, title, description, location, start_at, end_at, created_at, updated_at
		FROM events
		WHERE user_id = $1
		ORDER BY start_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Events
	for rows.Next() {
		var e models.Events
		err := rows.Scan(
			&e.ID,
			&e.UserID,
			&e.Title,
			&e.Description,
			&e.Location,
			&e.StartAt,
			&e.EndAt,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}