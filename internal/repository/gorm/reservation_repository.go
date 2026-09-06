package gormrepo

import (
	"restaurant-table-reservation/internal/domain"

	"gorm.io/gorm"
)

type ReservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(reservation *domain.Reservation) error {
	return r.db.Create(reservation).Error
}

func (r *ReservationRepository) FindByID(id uint) (*domain.Reservation, error) {
	var reservation domain.Reservation
	err := r.db.Preload("Table").First(&reservation, id).Error
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepository) Update(reservation *domain.Reservation) error {
	return r.db.Save(reservation).Error
}

func (r *ReservationRepository) FindConfirmedByTableDateSlot(tableID uint, date string, slot string) (*domain.Reservation, error) {
	var reservation domain.Reservation
	err := r.db.
		Where("table_id = ? AND reservation_date = ? AND time_slot = ? AND status = ?",
			tableID, date, slot, domain.ReservationStatusConfirmed).
		First(&reservation).Error
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepository) FindBookedTableIDs(date string, slot string) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&domain.Reservation{}).
		Where("reservation_date = ? AND time_slot = ? AND status = ?",
			date, slot, domain.ReservationStatusConfirmed).
		Pluck("table_id", &ids).Error
	return ids, err
}
