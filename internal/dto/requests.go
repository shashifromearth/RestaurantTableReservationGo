package dto

import (
	"time"

	"restaurant-table-reservation/internal/domain"
)

type CreateTableRequest struct {
	TableNumber string          `json:"table_number" binding:"required"`
	Capacity    int             `json:"capacity" binding:"required,min=1"`
	Location    domain.Location `json:"location" binding:"required"`
}

type TableResponse struct {
	ID          uint            `json:"id"`
	TableNumber string          `json:"table_number"`
	Capacity    int             `json:"capacity"`
	Location    domain.Location `json:"location"`
	CreatedAt   time.Time       `json:"created_at"`
}

type AvailableSlotResponse struct {
	TimeSlot         string          `json:"time_slot"`
	AvailableTables  []TableResponse `json:"available_tables"`
}

type AvailableSlotsResponse struct {
	Date  string                  `json:"date"`
	Slots []AvailableSlotResponse `json:"slots"`
}

type BookReservationRequest struct {
	TableID         *uint  `json:"table_id"`
	CustomerName    string `json:"customer_name" binding:"required"`
	CustomerEmail   string `json:"customer_email" binding:"required,email"`
	CustomerPhone   string `json:"customer_phone" binding:"required"`
	GuestCount      int    `json:"guest_count" binding:"required,min=1"`
	ReservationDate string `json:"reservation_date" binding:"required"`
	TimeSlot        string `json:"time_slot" binding:"required"`
	SpecialRequests string `json:"special_requests"`
}

type ReservationResponse struct {
	ID              uint                      `json:"id"`
	Table           TableResponse             `json:"table"`
	CustomerName    string                    `json:"customer_name"`
	CustomerEmail   string                    `json:"customer_email"`
	CustomerPhone   string                    `json:"customer_phone"`
	GuestCount      int                       `json:"guest_count"`
	ReservationDate string                    `json:"reservation_date"`
	TimeSlot        string                    `json:"time_slot"`
	SpecialRequests string                    `json:"special_requests"`
	Status          domain.ReservationStatus  `json:"status"`
	CreatedAt       time.Time                 `json:"created_at"`
}

type CancelReservationResponse struct {
	ID     uint                     `json:"id"`
	Status domain.ReservationStatus `json:"status"`
}

func ToTableResponse(t domain.Table) TableResponse {
	return TableResponse{
		ID:          t.ID,
		TableNumber: t.TableNumber,
		Capacity:    t.Capacity,
		Location:    t.Location,
		CreatedAt:   t.CreatedAt,
	}
}

func ToReservationResponse(r domain.Reservation) ReservationResponse {
	return ReservationResponse{
		ID:              r.ID,
		Table:           ToTableResponse(r.Table),
		CustomerName:    r.CustomerName,
		CustomerEmail:   r.CustomerEmail,
		CustomerPhone:   r.CustomerPhone,
		GuestCount:      r.GuestCount,
		ReservationDate: r.ReservationDate.Format("2006-01-02"),
		TimeSlot:        r.TimeSlot,
		SpecialRequests: r.SpecialRequests,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt,
	}
}
