package repository

import "restaurant-table-reservation/internal/domain"

type TableRepository interface {
	Create(table *domain.Table) error
	FindAll() ([]domain.Table, error)
	FindByID(id uint) (*domain.Table, error)
	FindAvailableForSlot(date string, slot string, minCapacity int) ([]domain.Table, error)
}
