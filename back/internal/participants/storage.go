package participants

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

// Create создает нового участника
func (r *Repo) Create(ctx context.Context, eventID int64, req models.ParticipantCreate) (*models.Participant, error) {
    var participant models.Participant
    
    query := `
        INSERT INTO participants (event_id, name, email, created_at)
        VALUES ($1, $2, $3, NOW())
        RETURNING id, event_id, name, email, created_at
    `

    err := r.db.QueryRow(ctx, query, eventID, req.Name, req.Email).Scan(
        &participant.ID,
        &participant.EventID,
        &participant.Name,
        &participant.Email,
        &participant.CreatedAt,
    )

    if err != nil {
        return nil, err
    }

    return &participant, nil
}

// List возвращает список участников события с пагинацией
func (r *Repo) List(ctx context.Context, eventID int64, page, limit int) ([]models.Participant, int, error) {
    offset := (page - 1) * limit

    // Получаем общее количество
    var total int
    countQuery := `SELECT COUNT(*) FROM participants WHERE event_id = $1`
    err := r.db.QueryRow(ctx, countQuery, eventID).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    // Получаем участников
    query := `
        SELECT id, event_id, name, email, created_at
        FROM participants
        WHERE event_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

    rows, err := r.db.Query(ctx, query, eventID, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var participants []models.Participant
    for rows.Next() {
        var p models.Participant
        err := rows.Scan(
            &p.ID,
            &p.EventID,
            &p.Name,
            &p.Email,
            &p.CreatedAt,
        )
        if err != nil {
            return nil, 0, err
        }
        participants = append(participants, p)
    }

    return participants, total, nil
}

// Delete удаляет участника
func (r *Repo) Delete(ctx context.Context, eventID, participantID int64) error {
    cmd, err := r.db.Exec(ctx, `DELETE FROM participants WHERE id = $1 AND event_id = $2`, participantID, eventID)
    if err != nil {
        return err
    }

    if cmd.RowsAffected() == 0 {
        return errors.New("participant not found")
    }

    return nil
}

// GetByID получает участника по ID (для проверки)
func (r *Repo) GetByID(ctx context.Context, id int64) (*models.Participant, error) {
    var p models.Participant
    query := `
        SELECT id, event_id, name, email, created_at
        FROM participants
        WHERE id = $1
    `

    err := r.db.QueryRow(ctx, query, id).Scan(
        &p.ID,
        &p.EventID,
        &p.Name,
        &p.Email,
        &p.CreatedAt,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("participant not found")
        }
        return nil, err
    }

    return &p, nil
}