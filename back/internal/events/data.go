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

// 2. GET BY ID - Получение события по ID (только не удаленные)
func (r *Repo) RGetEventsById(ctx context.Context, id int64) (*models.Events, error) {
	var events models.Events
	query := `SELECT id, user_id, title, description, location, start_at, end_at, created_at, updated_at, deleted_at 
	          FROM events WHERE id = $1 AND deleted_at IS NULL`  // ИЗМЕНЕНО: добавили фильтр deleted_at
	
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
		&events.DeletedAt,  // ДОБАВЛЕНО
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
		WHERE id = $6 AND deleted_at IS NULL  // ИЗМЕНЕНО: добавили проверку на deleted_at
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

// 4. DELETE - Мягкое удаление (soft delete) - ИЗМЕНЕНО
func (r *Repo) RDeleteEvents(ctx context.Context, id int64) error {
	cmd, err := r.db.Exec(ctx, `UPDATE events SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("event not found")
	}
	return nil
}

// 5. RESTORE - Восстановление события - НОВЫЙ МЕТОД
func (r *Repo) RRestoreEvent(ctx context.Context, id int64) error {
	cmd, err := r.db.Exec(ctx, `UPDATE events SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("event not found or not deleted")
	}
	return nil
}

// 6. LIST - Список событий с фильтрацией и пагинацией (только не удаленные)
func (r *Repo) RListEvents(ctx context.Context, userID int64, filter models.EventFilter) ([]models.Events, int, error) {
	offset := (filter.Page - 1) * filter.Limit

	// Базовый запрос - ИЗМЕНЕНО: добавили deleted_at IS NULL
	baseQuery := `FROM events WHERE user_id = $1 AND deleted_at IS NULL`
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

	// Получаем события - ИЗМЕНЕНО: добавили deleted_at в SELECT
	query := `
		SELECT id, user_id, title, description, location, start_at, end_at, created_at, updated_at, deleted_at
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
			&e.DeletedAt,  // ДОБАВЛЕНО
		)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}

	return events, total, nil
}

// Вспомогательный метод для проверки дубликатов (только не удаленные)
func (r *Repo) RCheckDuplicate(ctx context.Context, userID int64, title string, startAt time.Time) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM events WHERE user_id = $1 AND title = $2 AND start_at = $3 AND deleted_at IS NULL)`
	err := r.db.QueryRow(ctx, query, userID, title, startAt).Scan(&exists)
	return exists, err
}

// Получение событий по пользователю (только не удаленные) - ИЗМЕНЕНО
func (r *Repo) RGetEventsByUser(ctx context.Context, userID int64) ([]models.Events, error) {
	query := `
		SELECT id, user_id, title, description, location, start_at, end_at, created_at, updated_at, deleted_at
		FROM events
		WHERE user_id = $1 AND deleted_at IS NULL
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
			&e.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}
// GetEventStatistics - статистика событий за период
func (r *Repo) GetEventStatistics(ctx context.Context, userID int64, from, to time.Time) (map[string]interface{}, error) {
    // Общее количество событий
    var totalEvents int
    eventsQuery := `SELECT COUNT(*) FROM events 
                    WHERE user_id = $1 
                    AND start_at BETWEEN $2 AND $3 
                    AND deleted_at IS NULL`
    err := r.db.QueryRow(ctx, eventsQuery, userID, from, to).Scan(&totalEvents)
    if err != nil {
        return nil, err
    }

    // Общее количество участников
    var totalParticipants int
    participantsQuery := `
        SELECT COUNT(DISTINCT p.id) 
        FROM participants p
        JOIN events e ON p.event_id = e.id
        WHERE e.user_id = $1 
        AND e.start_at BETWEEN $2 AND $3
        AND e.deleted_at IS NULL`
    err = r.db.QueryRow(ctx, participantsQuery, userID, from, to).Scan(&totalParticipants)
    if err != nil {
        return nil, err
    }

    // События по дням (опционально)
    eventsByDayQuery := `
        SELECT DATE(start_at) as day, COUNT(*) as count
        FROM events
        WHERE user_id = $1 
        AND start_at BETWEEN $2 AND $3
        AND deleted_at IS NULL
        GROUP BY DATE(start_at)
        ORDER BY day`
    
    rows, err := r.db.Query(ctx, eventsByDayQuery, userID, from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    eventsByDay := make(map[string]int)
    for rows.Next() {
        var day time.Time
        var count int
        if err := rows.Scan(&day, &count); err != nil {
            return nil, err
        }
        eventsByDay[day.Format("2006-01-02")] = count
    }

    return map[string]interface{}{
        "total_events":       totalEvents,
        "total_participants": totalParticipants,
        "events_by_day":      eventsByDay,
    }, nil
}
// GetUpcomingEvents - ближайшие события пользователя
func (r *Repo) GetUpcomingEvents(ctx context.Context, userID int64, limit int) ([]models.Events, error) {
    query := `
        SELECT id, user_id, title, description, location, start_at, end_at, created_at, updated_at, deleted_at
        FROM events
        WHERE user_id = $1 
        AND start_at > NOW()
        AND deleted_at IS NULL
        ORDER BY start_at ASC
        LIMIT $2`

    rows, err := r.db.Query(ctx, query, userID, limit)
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
            &e.DeletedAt,
        )
        if err != nil {
            return nil, err
        }
        events = append(events, e)
    }

    return events, nil
}