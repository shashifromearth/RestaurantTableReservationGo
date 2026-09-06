package repository

import "restaurant-table-reservation/internal/domain"

type ReservationRepository interface {
	Create(reservation *domain.Reservation) error
	FindByID(id uint) (*domain.Reservation, error)
	Update(reservation *domain.Reservation) error
	FindConfirmedByTableDateSlot(tableID uint, date string, slot string) (*domain.Reservation, error)
	FindBookedTableIDs(date string, slot string) ([]uint, error)
}
