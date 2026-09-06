package main

import (
	"log"

	"restaurant-table-reservation/internal/config"
	"restaurant-table-reservation/internal/database"
	"restaurant-table-reservation/internal/handler"
	gormrepo "restaurant-table-reservation/internal/repository/gorm"
	"restaurant-table-reservation/internal/router"
	"restaurant-table-reservation/internal/service"

	_ "restaurant-table-reservation/docs"
)

// @title           Restaurant Table Reservation API
// @version         1.0
// @description     REST API for restaurant table reservations with smart table assignment.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	tableRepo := gormrepo.NewTableRepository(db)
	reservationRepo := gormrepo.NewReservationRepository(db)

	tableService := service.NewTableService(tableRepo)
	reservationService := service.NewReservationService(tableRepo, reservationRepo, cfg.CancellationLeadHours)

	tableHandler := handler.NewTableHandler(tableService)
	reservationHandler := handler.NewReservationHandler(reservationService)

	engine := router.Setup(tableHandler, reservationHandler)

	log.Printf("server starting on :%s", cfg.Port)
	if err := engine.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
