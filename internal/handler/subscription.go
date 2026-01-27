package handler

import (
	"errors"
	"net/http"

	"em_tz_anvar/internal/models"
	"em_tz_anvar/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// CreateSubscription создает новую подписку
// @Summary Создание подписки
// @Description Создает новую запись о подписке пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body model.CreateSubscriptionRequest true "Данные подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	subscription, err := h.services.Subscription.Create(c.Request.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create subscription")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription возвращает подписку по ID
// @Summary Получение подписки
// @Description Возвращает подписку по её ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Warn().Err(err).Str("id", c.Param("id")).Msg("Invalid subscription ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	subscription, err := h.services.Subscription.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
			return
		}
		log.Error().Err(err).Str("subscription_id", id.String()).Msg("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// GetAllSubscriptions возвращает список подписок
// @Summary Список подписок
// @Description Возвращает список всех подписок с возможностью фильтрации
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя (UUID)"
// @Param service_name query string false "Название сервиса"
// @Param limit query int false "Лимит записей" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *Handler) GetAllSubscriptions(c *gin.Context) {
	filter := &models.SubscriptionFilter{
		ServiceName: c.Query("service_name"),
		Limit:       20,
		Offset:      0,
	}

	// Парсинг user_id
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userIDStr).Msg("Invalid user ID")
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id format"})
			return
		}
		filter.UserID = &userID
	}

	// Парсинг limit и offset
	if limit := c.Query("limit"); limit != "" {
		var l int
		if _, err := parseQueryInt(limit, &l); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		var o int
		if _, err := parseQueryInt(offset, &o); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	subscriptions, err := h.services.Subscription.GetAll(c.Request.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscriptions")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// UpdateSubscription обновляет подписку
// @Summary Обновление подписки
// @Description Обновляет существующую подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Param input body model.UpdateSubscriptionRequest true "Данные для обновления"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Warn().Err(err).Str("id", c.Param("id")).Msg("Invalid subscription ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	var req models.UpdateSubscriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	subscription, err := h.services.Subscription.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
			return
		}
		log.Error().Err(err).Str("subscription_id", id.String()).Msg("Failed to update subscription")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription удаляет подписку
// @Summary Удаление подписки
// @Description Удаляет подписку по ID
// @Tags subscriptions
// @Param id path string true "ID подписки (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Warn().Err(err).Str("id", c.Param("id")).Msg("Invalid subscription ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	err = h.services.Subscription.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
			return
		}
		log.Error().Err(err).Str("subscription_id", id.String()).Msg("Failed to delete subscription")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTotalCost возвращает суммарную стоимость подписок за период
// @Summary Суммарная стоимость подписок
// @Description Подсчитывает суммарную стоимость всех подписок за выбранный период
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя (UUID)"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string true "Начало периода (MM-YYYY)"
// @Param end_date query string true "Конец периода (MM-YYYY)"
// @Success 200 {object} model.TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/cost [get]
func (h *Handler) GetTotalCost(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "start_date and end_date are required"})
		return
	}

	startDate, err := parseMonthYear(startDateStr)
	if err != nil {
		log.Warn().Err(err).Str("start_date", startDateStr).Msg("Invalid start_date format")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format, expected MM-YYYY"})
		return
	}

	endDate, err := parseMonthYear(endDateStr)
	if err != nil {
		log.Warn().Err(err).Str("end_date", endDateStr).Msg("Invalid end_date format")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid end_date format, expected MM-YYYY"})
		return
	}

	filter := &models.CostFilter{
		StartDate:   startDate,
		EndDate:     endDate,
		ServiceName: c.Query("service_name"),
	}

	// Парсинг user_id
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userIDStr).Msg("Invalid user ID")
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id format"})
			return
		}
		filter.UserID = &userID
	}

	result, err := h.services.Subscription.GetTotalCost(c.Request.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to calculate total cost")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
