package handler

import (
	"errors"
	"net/http"
	"strconv"

	"restaurant-table-reservation/internal/apperror"
	"restaurant-table-reservation/internal/dto"
	"restaurant-table-reservation/internal/service"

	"github.com/gin-gonic/gin"
)

type TableHandler struct {
	tableService *service.TableService
}

func NewTableHandler(tableService *service.TableService) *TableHandler {
	return &TableHandler{tableService: tableService}
}

// CreateTable godoc
// @Summary      Create a table
// @Description  Restaurant staff adds a new table
// @Tags         tables
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateTableRequest true "Table details"
// @Success      201 {object} dto.TableResponse
// @Failure      400 {object} apperror.AppError
// @Failure      409 {object} apperror.AppError
// @Router       /tables [post]
func (h *TableHandler) CreateTable(c *gin.Context) {
	var req dto.CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, apperror.From(apperror.ErrInvalidInput))
		return
	}

	table, err := h.tableService.CreateTable(req)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, table)
}

// ListTables godoc
// @Summary      List all tables
// @Tags         tables
// @Produce      json
// @Success      200 {array} dto.TableResponse
// @Router       /tables [get]
func (h *TableHandler) ListTables(c *gin.Context) {
	tables, err := h.tableService.ListTables()
	if err != nil {
		respondError(c, http.StatusInternalServerError, apperror.From(err))
		return
	}
	c.JSON(http.StatusOK, tables)
}

type ReservationHandler struct {
	reservationService *service.ReservationService
}

func NewReservationHandler(reservationService *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{reservationService: reservationService}
}

// GetAvailableSlots godoc
// @Summary      View available slots
// @Description  Shows available time slots and tables for a date
// @Tags         reservations
// @Produce      json
// @Param        date query string true "Date (YYYY-MM-DD)"
// @Success      200 {object} dto.AvailableSlotsResponse
// @Failure      400 {object} apperror.AppError
// @Router       /slots [get]
func (h *ReservationHandler) GetAvailableSlots(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		respondError(c, http.StatusBadRequest, apperror.New("INVALID_INPUT", "date query parameter is required"))
		return
	}

	slots, err := h.reservationService.GetAvailableSlots(date)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, slots)
}

// BookReservation godoc
// @Summary      Book a table
// @Description  Create a reservation. Omit table_id for smart table assignment.
// @Tags         reservations
// @Accept       json
// @Produce      json
// @Param        request body dto.BookReservationRequest true "Reservation details"
// @Success      201 {object} dto.ReservationResponse
// @Failure      400 {object} apperror.AppError
// @Failure      404 {object} apperror.AppError
// @Failure      409 {object} apperror.AppError
// @Router       /reservations [post]
func (h *ReservationHandler) BookReservation(c *gin.Context) {
	var req dto.BookReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, apperror.From(apperror.ErrInvalidInput))
		return
	}

	reservation, err := h.reservationService.Book(req)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

// GetReservation godoc
// @Summary      View a reservation
// @Tags         reservations
// @Produce      json
// @Param        id path int true "Reservation ID"
// @Success      200 {object} dto.ReservationResponse
// @Failure      404 {object} apperror.AppError
// @Router       /reservations/{id} [get]
func (h *ReservationHandler) GetReservation(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, apperror.From(apperror.ErrInvalidInput))
		return
	}

	reservation, err := h.reservationService.GetByID(id)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, reservation)
}

// CancelReservation godoc
// @Summary      Cancel a reservation
// @Description  Cancellation must occur before the configured lead time
// @Tags         reservations
// @Produce      json
// @Param        id path int true "Reservation ID"
// @Success      200 {object} dto.CancelReservationResponse
// @Failure      400 {object} apperror.AppError
// @Failure      404 {object} apperror.AppError
// @Router       /reservations/{id} [delete]
func (h *ReservationHandler) CancelReservation(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, apperror.From(apperror.ErrInvalidInput))
		return
	}

	result, err := h.reservationService.Cancel(id)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, apperror.ErrInvalidInput
	}
	return uint(id), nil
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		respondError(c, http.StatusNotFound, apperror.From(err))
	case errors.Is(err, apperror.ErrInvalidInput):
		respondError(c, http.StatusBadRequest, apperror.From(err))
	case errors.Is(err, apperror.ErrTableAlreadyBooked),
		errors.Is(err, apperror.ErrDuplicateTable):
		respondError(c, http.StatusConflict, apperror.From(err))
	case errors.Is(err, apperror.ErrCapacityExceeded),
		errors.Is(err, apperror.ErrCancellationDenied),
		errors.Is(err, apperror.ErrAlreadyCancelled),
		errors.Is(err, apperror.ErrNoAvailableTable):
		respondError(c, http.StatusBadRequest, apperror.From(err))
	default:
		respondError(c, http.StatusInternalServerError, apperror.From(err))
	}
}

func respondError(c *gin.Context, status int, body apperror.AppError) {
	c.JSON(status, body)
}
