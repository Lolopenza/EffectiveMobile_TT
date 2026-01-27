package handler

import (
	"em_tz_anvar/internal/service"

	_ "em_tz_anvar/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	services *service.Service
}

// NewHandler
func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

// InitRoutes
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware())

	//Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", h.CreateSubscription)
			subscriptions.GET("", h.GetAllSubscriptions)
			// Эндпоинт для подсчета стоимости (должен быть перед /:id)
			subscriptions.GET("/cost", h.GetTotalCost)
			subscriptions.GET("/:id", h.GetSubscription)
			subscriptions.PUT("/:id", h.UpdateSubscription)
			subscriptions.DELETE("/:id", h.DeleteSubscription)
		}
	}

	//Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}
