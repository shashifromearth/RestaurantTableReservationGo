package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"restaurant-table-reservation/internal/database"
	"restaurant-table-reservation/internal/domain"
	"restaurant-table-reservation/internal/dto"
	"restaurant-table-reservation/internal/handler"
	gormrepo "restaurant-table-reservation/internal/repository/gorm"
	"restaurant-table-reservation/internal/router"
	"restaurant-table-reservation/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	db, err := database.Connect(t.TempDir() + "/test.db")
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	tableRepo := gormrepo.NewTableRepository(db)
	reservationRepo := gormrepo.NewReservationRepository(db)

	tableService := service.NewTableService(tableRepo)
	reservationService := service.NewReservationService(tableRepo, reservationRepo, 2)

	engine := router.Setup(
		handler.NewTableHandler(tableService),
		handler.NewReservationHandler(reservationService),
	)

	return engine
}

func TestAPI_EndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := setupTestRouter(t)

	createBody, _ := json.Marshal(dto.CreateTableRequest{
		TableNumber: "A1",
		Capacity:    4,
		Location:    domain.LocationIndoor,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tables", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	bookBody, _ := json.Marshal(dto.BookReservationRequest{
		CustomerName:    "John Doe",
		CustomerEmail:   "john@example.com",
		CustomerPhone:   "555-1234",
		GuestCount:      2,
		ReservationDate: "2026-12-01",
		TimeSlot:        "19:00",
		SpecialRequests: "Window seat",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/reservations", bytes.NewReader(bookBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var reservation dto.ReservationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &reservation))
	require.NotZero(t, reservation.ID)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/reservations/"+strconv.FormatUint(uint64(reservation.ID), 10), nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/slots?date=2026-12-01", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/reservations/"+strconv.FormatUint(uint64(reservation.ID), 10), nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
