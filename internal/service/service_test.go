package service_test

import (
	"fmt"
	"testing"
	"time"

	"restaurant-table-reservation/internal/apperror"
	"restaurant-table-reservation/internal/domain"
	"restaurant-table-reservation/internal/dto"
	"restaurant-table-reservation/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockTableRepo struct {
	tables      []domain.Table
	available   []domain.Table
	createErr   error
	findByID    *domain.Table
	findByIDErr error
}

func (m *mockTableRepo) Create(table *domain.Table) error {
	if m.createErr != nil {
		return m.createErr
	}
	table.ID = uint(len(m.tables) + 1)
	m.tables = append(m.tables, *table)
	return nil
}

func (m *mockTableRepo) FindAll() ([]domain.Table, error) {
	return m.tables, nil
}

func (m *mockTableRepo) FindByID(id uint) (*domain.Table, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if m.findByID != nil {
		return m.findByID, nil
	}
	for _, t := range m.tables {
		if t.ID == id {
			copy := t
			return &copy, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockTableRepo) FindAvailableForSlot(date string, slot string, minCapacity int) ([]domain.Table, error) {
	return m.available, nil
}

type mockReservationRepo struct {
	reservations map[uint]domain.Reservation
	nextID       uint
	booked       map[string]uint
}

func newMockReservationRepo() *mockReservationRepo {
	return &mockReservationRepo{
		reservations: make(map[uint]domain.Reservation),
		booked:       make(map[string]uint),
	}
}

func (m *mockReservationRepo) key(tableID uint, date, slot string) string {
	return fmtKey(tableID, date, slot)
}

func fmtKey(tableID uint, date, slot string) string {
	return fmt.Sprintf("%d|%s|%s", tableID, date, slot)
}

func (m *mockReservationRepo) Create(reservation *domain.Reservation) error {
	m.nextID++
	reservation.ID = m.nextID
	m.reservations[reservation.ID] = *reservation
	m.booked[m.key(reservation.TableID, reservation.ReservationDate.Format("2006-01-02"), reservation.TimeSlot)] = reservation.ID
	return nil
}

func (m *mockReservationRepo) FindByID(id uint) (*domain.Reservation, error) {
	r, ok := m.reservations[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &r, nil
}

func (m *mockReservationRepo) Update(reservation *domain.Reservation) error {
	m.reservations[reservation.ID] = *reservation
	return nil
}

func (m *mockReservationRepo) FindConfirmedByTableDateSlot(tableID uint, date string, slot string) (*domain.Reservation, error) {
	if _, ok := m.booked[m.key(tableID, date, slot)]; ok {
		for _, r := range m.reservations {
			if r.TableID == tableID && r.ReservationDate.Format("2006-01-02") == date && r.TimeSlot == slot && r.Status == domain.ReservationStatusConfirmed {
				copy := r
				return &copy, nil
			}
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockReservationRepo) FindBookedTableIDs(date string, slot string) ([]uint, error) {
	var ids []uint
	for _, r := range m.reservations {
		if r.ReservationDate.Format("2006-01-02") == date && r.TimeSlot == slot && r.Status == domain.ReservationStatusConfirmed {
			ids = append(ids, r.TableID)
		}
	}
	return ids, nil
}

func TestTableService_CreateTable(t *testing.T) {
	repo := &mockTableRepo{}
	svc := service.NewTableService(repo)

	resp, err := svc.CreateTable(dto.CreateTableRequest{
		TableNumber: "T1",
		Capacity:    4,
		Location:    domain.LocationIndoor,
	})

	require.NoError(t, err)
	assert.Equal(t, "T1", resp.TableNumber)
	assert.Equal(t, 4, resp.Capacity)
}

func TestTableService_InvalidLocation(t *testing.T) {
	repo := &mockTableRepo{}
	svc := service.NewTableService(repo)

	_, err := svc.CreateTable(dto.CreateTableRequest{
		TableNumber: "T1",
		Capacity:    4,
		Location:    "patio",
	})

	assert.ErrorIs(t, err, apperror.ErrInvalidInput)
}

func TestReservationService_SmartAssignment(t *testing.T) {
	tableRepo := &mockTableRepo{
		available: []domain.Table{
			{ID: 1, TableNumber: "T1", Capacity: 2},
			{ID: 2, TableNumber: "T2", Capacity: 6},
		},
	}
	resRepo := newMockReservationRepo()
	svc := service.NewReservationService(tableRepo, resRepo, 2)

	resp, err := svc.Book(dto.BookReservationRequest{
		CustomerName:    "Jane",
		CustomerEmail:   "jane@example.com",
		CustomerPhone:   "555-0100",
		GuestCount:      2,
		ReservationDate: "2026-09-10",
		TimeSlot:        "19:00",
	})

	require.NoError(t, err)
	assert.Equal(t, uint(1), resp.Table.ID)
	assert.Equal(t, 2, resp.Table.Capacity)
}

func TestReservationService_CapacityExceeded(t *testing.T) {
	tableRepo := &mockTableRepo{
		findByID: &domain.Table{ID: 1, Capacity: 2},
	}
	resRepo := newMockReservationRepo()
	svc := service.NewReservationService(tableRepo, resRepo, 2)

	tableID := uint(1)
	_, err := svc.Book(dto.BookReservationRequest{
		TableID:         &tableID,
		CustomerName:    "Jane",
		CustomerEmail:   "jane@example.com",
		CustomerPhone:   "555-0100",
		GuestCount:      5,
		ReservationDate: "2026-09-10",
		TimeSlot:        "19:00",
	})

	assert.ErrorIs(t, err, apperror.ErrCapacityExceeded)
}

func TestReservationService_DoubleBookingPrevented(t *testing.T) {
	tableRepo := &mockTableRepo{
		findByID: &domain.Table{ID: 1, Capacity: 4},
	}
	resRepo := newMockReservationRepo()
	svc := service.NewReservationService(tableRepo, resRepo, 2)

	tableID := uint(1)
	req := dto.BookReservationRequest{
		TableID:         &tableID,
		CustomerName:    "Jane",
		CustomerEmail:   "jane@example.com",
		CustomerPhone:   "555-0100",
		GuestCount:      2,
		ReservationDate: "2026-09-10",
		TimeSlot:        "19:00",
	}

	_, err := svc.Book(req)
	require.NoError(t, err)

	_, err = svc.Book(req)
	assert.ErrorIs(t, err, apperror.ErrTableAlreadyBooked)
}

func TestReservationService_CancelWithinLeadTimeDenied(t *testing.T) {
	tableRepo := &mockTableRepo{}
	resRepo := newMockReservationRepo()
	svc := service.NewReservationService(tableRepo, resRepo, 2)
	svcNow := time.Date(2026, 9, 10, 18, 30, 0, 0, time.UTC)
	svc.SetNowFunc(func() time.Time { return svcNow })

	date := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	resRepo.reservations[1] = domain.Reservation{
		ID:              1,
		TableID:         1,
		ReservationDate: date,
		TimeSlot:        "19:00",
		Status:          domain.ReservationStatusConfirmed,
	}

	_, err := svc.Cancel(1)
	assert.ErrorIs(t, err, apperror.ErrCancellationDenied)
}

func TestReservationService_CancelSuccess(t *testing.T) {
	tableRepo := &mockTableRepo{}
	resRepo := newMockReservationRepo()
	svc := service.NewReservationService(tableRepo, resRepo, 2)
	svcNow := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	svc.SetNowFunc(func() time.Time { return svcNow })

	date := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	resRepo.reservations[1] = domain.Reservation{
		ID:              1,
		TableID:         1,
		ReservationDate: date,
		TimeSlot:        "19:00",
		Status:          domain.ReservationStatusConfirmed,
	}

	resp, err := svc.Cancel(1)
	require.NoError(t, err)
	assert.Equal(t, domain.ReservationStatusCancelled, resp.Status)
}
