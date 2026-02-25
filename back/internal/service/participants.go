package service

import (
    "context"
    "errors"  // добавить
    "net/http"
    "search-job/internal/models"
    "strconv"

    "github.com/labstack/echo/v4"
)

// AddParticipant - добавить участника к событию
func (s *Service) AddParticipant(c echo.Context) error {
    // Проверяем авторизацию
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }

    // Получаем ID события из URL
    eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid event ID"))
    }

    // TODO: Проверить, что событие принадлежит пользователю
    // Для этого нужен доступ к eventsRepo

    var req models.ParticipantCreate
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid request format"))
    }

    participant, err := s.addParticipant(c.Request().Context(), eventID, req)
    if err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse(err.Error()))
    }

    return c.JSON(http.StatusCreated, s.SuccessResponse(participant))
}

// ListParticipants - список участников события
func (s *Service) ListParticipants(c echo.Context) error {
    // Проверяем авторизацию
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }

    // Получаем ID события из URL
    eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid event ID"))
    }

    // Парсинг пагинации
    page, _ := strconv.Atoi(c.QueryParam("page"))
    if page < 1 {
        page = 1
    }

    limit, _ := strconv.Atoi(c.QueryParam("limit"))
    if limit < 1 {
        limit = 10
    }

    participants, total, err := s.listParticipants(c.Request().Context(), eventID, page, limit)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("failed to fetch participants"))
    }

    return c.JSON(http.StatusOK, s.ListResponse(participants, total, page, limit))
}

// RemoveParticipant - удалить участника
func (s *Service) RemoveParticipant(c echo.Context) error {
    // Проверяем авторизацию
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }

    // Получаем ID события из URL
    eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid event ID"))
    }

    // Получаем ID участника из URL
    participantID, err := strconv.ParseInt(c.Param("participantId"), 10, 64)
    if err != nil {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid participant ID"))
    }

    err = s.removeParticipant(c.Request().Context(), eventID, participantID)
    if err != nil {
        return c.JSON(http.StatusNotFound, s.ErrorResponse(err.Error()))
    }

    return c.JSON(http.StatusNoContent, nil)
}

// Внутренние методы (бизнес-логика)

func (s *Service) addParticipant(ctx context.Context, eventID int64, req models.ParticipantCreate) (*models.Participant, error) {
    if s.participantRepo == nil {
        return nil, errors.New("participant repo not initialized")
    }
    return s.participantRepo.Create(ctx, eventID, req)
}

func (s *Service) listParticipants(ctx context.Context, eventID int64, page, limit int) ([]models.Participant, int, error) {
    if s.participantRepo == nil {
        return nil, 0, errors.New("participant repo not initialized")
    }
    return s.participantRepo.List(ctx, eventID, page, limit)
}

func (s *Service) removeParticipant(ctx context.Context, eventID, participantID int64) error {
    if s.participantRepo == nil {
        return errors.New("participant repo not initialized")
    }
    return s.participantRepo.Delete(ctx, eventID, participantID)
}