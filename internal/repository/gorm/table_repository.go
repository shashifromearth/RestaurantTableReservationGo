package gormrepo

import (
	"restaurant-table-reservation/internal/domain"

	"gorm.io/gorm"
)

type TableRepository struct {
	db *gorm.DB
}

func NewTableRepository(db *gorm.DB) *TableRepository {
	return &TableRepository{db: db}
}

func (r *TableRepository) Create(table *domain.Table) error {
	return r.db.Create(table).Error
}

func (r *TableRepository) FindAll() ([]domain.Table, error) {
	var tables []domain.Table
	err := r.db.Order("table_number asc").Find(&tables).Error
	return tables, err
}

func (r *TableRepository) FindByID(id uint) (*domain.Table, error) {
	var table domain.Table
	err := r.db.First(&table, id).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *TableRepository) FindAvailableForSlot(date string, slot string, minCapacity int) ([]domain.Table, error) {
	var tables []domain.Table

	subQuery := r.db.Model(&domain.Reservation{}).
		Select("table_id").
		Where("reservation_date = ? AND time_slot = ? AND status = ?",
			date, slot, domain.ReservationStatusConfirmed)

	err := r.db.
		Where("capacity >= ?", minCapacity).
		Where("id NOT IN (?)", subQuery).
		Order("capacity asc, table_number asc").
		Find(&tables).Error

	return tables, err
}
