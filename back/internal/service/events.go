package service

import (
    "net/http"
    "search-job/internal/models"
    "strconv"
    "time"

    "github.com/labstack/echo/v4"
)

// 1. CREATE - Создание события
func (s *Service) CreateEvents(c echo.Context) error {
    // Получаем user_id напрямую из контекста (устанавливается middleware)
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }
    
    userID, ok := userIDVal.(int64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }

    var events models.Events
    if err := c.Bind(&events); err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid request format"))
    }

    // Устанавливаем user_id из токена
    events.UserID = userID

    // Валидация обязательных полей
    if events.Title == "" {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("title is required"))
    }
    if events.StartAt.IsZero() {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("startAt is required"))
    }

    // Валидация endAt >= startAt
    if !events.EndAt.IsZero() && events.EndAt.Before(events.StartAt) {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("endAt must be after or equal to startAt"))
    }

    // Проверка дубликатов
    repo := s.eventsRepo
    exists, err := repo.RCheckDuplicate(c.Request().Context(), events.UserID, events.Title, events.StartAt)
    if err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
    }
    if exists {
        return c.JSON(http.StatusConflict, s.ErrorResponse("Event with same title and start time already exists"))
    }

    if err := repo.RCreateEvents(c.Request().Context(), &events); err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
    }

    return c.JSON(http.StatusCreated, s.SuccessResponse(events))
}

// 2. GET BY ID - Получение события по ID
func (s *Service) GetEventsById(c echo.Context) error {
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }
    
    userID, ok := userIDVal.(int64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid event id"))
    }

    repo := s.eventsRepo

    event, err := repo.RGetEventsById(c.Request().Context(), id)
    if err != nil {
        if err.Error() == "event not found" {
            return c.JSON(http.StatusNotFound, s.ErrorResponse("event not found"))
        }
        s.logger.Error(err)
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
    }

    // Проверяем, что событие принадлежит пользователю
    if event.UserID != userID {
        s.logger.Errorf("user %d attempted to access event %d owned by %d", userID, id, event.UserID)
        return c.JSON(http.StatusForbidden, s.ErrorResponse("access denied"))
    }

    return c.JSON(http.StatusOK, s.SuccessResponse(event))
}

// 3. UPDATE - Обновление события
func (s *Service) UpdateEvents(c echo.Context) error {
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }
    
    userID, ok := userIDVal.(int64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid event id"))
    }

    var updateData models.UpdateEventRequest
    if err := c.Bind(&updateData); err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid request format"))
    }

    repo := s.eventsRepo

    // Получаем существующее событие
    event, err := repo.RGetEventsById(c.Request().Context(), id)
    if err != nil {
        if err.Error() == "event not found" {
            return c.JSON(http.StatusNotFound, s.ErrorResponse("event not found"))
        }
        s.logger.Error(err)
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
    }

    // Проверяем, что событие принадлежит пользователю
    if event.UserID != userID {
        s.logger.Errorf("user %d attempted to update event %d owned by %d", userID, id, event.UserID)
        return c.JSON(http.StatusForbidden, s.ErrorResponse("access denied"))
    }

    // Обновляем только те поля, которые были переданы
    if updateData.Title != nil {
        event.Title = *updateData.Title
    }
    if updateData.Description != nil {
        event.Description = *updateData.Description
    }
    if updateData.Location != nil {
        event.Location = *updateData.Location
    }
    if updateData.StartAt != nil {
        event.StartAt = *updateData.StartAt
    }
    if updateData.EndAt != nil {
        event.EndAt = *updateData.EndAt
    }

    // Валидация
    if !event.EndAt.IsZero() && event.EndAt.Before(event.StartAt) {
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("endAt must be after or equal to startAt"))
    }

    // Проверка дубликатов при изменении названия или времени
    if updateData.Title != nil || updateData.StartAt != nil {
        exists, err := repo.RCheckDuplicate(c.Request().Context(), event.UserID, event.Title, event.StartAt)
        if err != nil {
            s.logger.Error(err)
            return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
        }
        if exists {
            return c.JSON(http.StatusConflict, s.ErrorResponse("Event with same title and start time already exists"))
        }
    }

    if err := repo.RUpdateEvents(c.Request().Context(), event); err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
    }

    return c.JSON(http.StatusOK, s.SuccessResponse(event))
}

// 4. DELETE - Удаление события
func (s *Service) DeleteEvents(c echo.Context) error {
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }
    
    userID, ok := userIDVal.(int64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusBadRequest, s.ErrorResponse("invalid event id"))
    }

    repo := s.eventsRepo

    // Сначала получаем событие, чтобы проверить владельца
    event, err := repo.RGetEventsById(c.Request().Context(), id)
    if err != nil {
        if err.Error() == "event not found" {
            return c.JSON(http.StatusNotFound, s.ErrorResponse("event not found"))
        }
        s.logger.Error(err)
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
    }

    // Проверяем, что событие принадлежит пользователю
    if event.UserID != userID {
        s.logger.Errorf("user %d attempted to delete event %d owned by %d", userID, id, event.UserID)
        return c.JSON(http.StatusForbidden, s.ErrorResponse("access denied"))
    }

    if err := repo.RDeleteEvents(c.Request().Context(), id); err != nil {
        if err.Error() == "event not found" {
            return c.JSON(http.StatusNotFound, s.ErrorResponse("event not found"))
        }
        s.logger.Error(err)
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
    }

    return c.JSON(http.StatusNoContent, nil)
}

// 5. LIST - Список событий с фильтрацией и пагинацией
func (s *Service) ListEvents(c echo.Context) error {
    userIDVal := c.Get("user_id")
    if userIDVal == nil {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }
    
    userID, ok := userIDVal.(int64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, s.ErrorResponse("unauthorized"))
    }

    // Парсинг параметров фильтрации
    var filter models.EventFilter

    // from/to
    if fromStr := c.QueryParam("from"); fromStr != "" {
        from, err := time.Parse(time.RFC3339, fromStr)
        if err == nil {
            filter.From = &from
        }
    }
    if toStr := c.QueryParam("to"); toStr != "" {
        to, err := time.Parse(time.RFC3339, toStr)
        if err == nil {
            filter.To = &to
        }
    }

    // search
    if search := c.QueryParam("search"); search != "" {
        filter.Search = &search
    }

    // Пагинация
    page, err := strconv.Atoi(c.QueryParam("page"))
    if err != nil || page < 1 {
        page = 1
    }
    filter.Page = page

    limit, err := strconv.Atoi(c.QueryParam("limit"))
    if err != nil || limit < 1 {
        limit = 10
    }
    if limit > 100 {
        limit = 100
    }
    filter.Limit = limit

    // Сортировка
    filter.Sort = c.QueryParam("sort")
    if filter.Sort != "start_at" && filter.Sort != "created_at" {
        filter.Sort = "created_at"
    }

    filter.Order = c.QueryParam("order")
    if filter.Order != "asc" && filter.Order != "desc" {
        filter.Order = "desc"
    }

    repo := s.eventsRepo

    events, total, err := repo.RListEvents(c.Request().Context(), userID, filter)
    if err != nil {
        s.logger.Error(err)
        return c.JSON(http.StatusInternalServerError, s.ErrorResponse("internal error"))
    }

    return c.JSON(http.StatusOK, s.ListResponse(events, total, page, limit))
}