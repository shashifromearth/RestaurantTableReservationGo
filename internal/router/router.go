package router

import (
	"restaurant-table-reservation/internal/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(tableHandler *handler.TableHandler, reservationHandler *handler.ReservationHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/tables", tableHandler.CreateTable)
		api.GET("/tables", tableHandler.ListTables)

		api.GET("/slots", reservationHandler.GetAvailableSlots)
		api.POST("/reservations", reservationHandler.BookReservation)
		api.GET("/reservations/:id", reservationHandler.GetReservation)
		api.DELETE("/reservations/:id", reservationHandler.CancelReservation)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
