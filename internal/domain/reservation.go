package domain

import "time"

type ReservationStatus string

const (
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
)

type Reservation struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	TableID         uint              `gorm:"not null;index" json:"table_id"`
	Table           Table             `gorm:"foreignKey:TableID" json:"table,omitempty"`
	CustomerName    string            `gorm:"size:100;not null" json:"customer_name"`
	CustomerEmail   string            `gorm:"size:150;not null" json:"customer_email"`
	CustomerPhone   string            `gorm:"size:20;not null" json:"customer_phone"`
	GuestCount      int               `gorm:"not null" json:"guest_count"`
	ReservationDate time.Time         `gorm:"type:date;not null;index" json:"reservation_date"`
	TimeSlot        string            `gorm:"size:10;not null" json:"time_slot"`
	SpecialRequests string            `gorm:"size:500" json:"special_requests"`
	Status          ReservationStatus `gorm:"size:20;not null;default:confirmed" json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

func (Reservation) TableName() string {
	return "reservations"
}

func (s ReservationStatus) Valid() bool {
	return s == ReservationStatusConfirmed || s == ReservationStatusCancelled
}
