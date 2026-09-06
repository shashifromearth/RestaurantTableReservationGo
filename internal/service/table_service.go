package service

import (
	"errors"
	"fmt"

	"restaurant-table-reservation/internal/apperror"
	"restaurant-table-reservation/internal/domain"
	"restaurant-table-reservation/internal/dto"
	"restaurant-table-reservation/internal/repository"

	"gorm.io/gorm"
)

type TableService struct {
	tableRepo repository.TableRepository
}

func NewTableService(tableRepo repository.TableRepository) *TableService {
	return &TableService{tableRepo: tableRepo}
}

func (s *TableService) CreateTable(req dto.CreateTableRequest) (dto.TableResponse, error) {
	if !req.Location.Valid() {
		return dto.TableResponse{}, fmt.Errorf("%w: location must be indoor or outdoor", apperror.ErrInvalidInput)
	}

	table := &domain.Table{
		TableNumber: req.TableNumber,
		Capacity:    req.Capacity,
		Location:    req.Location,
	}

	if err := s.tableRepo.Create(table); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return dto.TableResponse{}, apperror.ErrDuplicateTable
		}
		return dto.TableResponse{}, err
	}

	return dto.ToTableResponse(*table), nil
}

func (s *TableService) ListTables() ([]dto.TableResponse, error) {
	tables, err := s.tableRepo.FindAll()
	if err != nil {
		return nil, err
	}

	result := make([]dto.TableResponse, len(tables))
	for i, t := range tables {
		result[i] = dto.ToTableResponse(t)
	}
	return result, nil
}
