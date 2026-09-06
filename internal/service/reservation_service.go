package service

import (
	"errors"
	"fmt"
	"time"

	"restaurant-table-reservation/internal/apperror"
	"restaurant-table-reservation/internal/domain"
	"restaurant-table-reservation/internal/dto"
	"restaurant-table-reservation/internal/repository"
	"restaurant-table-reservation/internal/timeslot"

	"gorm.io/gorm"
)

type ReservationService struct {
	tableRepo       repository.TableRepository
	reservationRepo repository.ReservationRepository
	cancelLeadHours int
	now             func() time.Time
}

func NewReservationService(
	tableRepo repository.TableRepository,
	reservationRepo repository.ReservationRepository,
	cancelLeadHours int,
) *ReservationService {
	return &ReservationService{
		tableRepo:       tableRepo,
		reservationRepo: reservationRepo,
		cancelLeadHours: cancelLeadHours,
		now:             time.Now,
	}
}

func (s *ReservationService) GetAvailableSlots(dateStr string) (dto.AvailableSlotsResponse, error) {
	date, err := parseDate(dateStr)
	if err != nil {
		return dto.AvailableSlotsResponse{}, err
	}

	tables, err := s.tableRepo.FindAll()
	if err != nil {
		return dto.AvailableSlotsResponse{}, err
	}

	slots := make([]dto.AvailableSlotResponse, 0, len(timeslot.All))
	for _, slot := range timeslot.All {
		bookedIDs, err := s.reservationRepo.FindBookedTableIDs(dateStr, slot)
		if err != nil {
			return dto.AvailableSlotsResponse{}, err
		}

		booked := toSet(bookedIDs)
		available := make([]dto.TableResponse, 0)
		for _, t := range tables {
			if !booked[t.ID] {
				available = append(available, dto.ToTableResponse(t))
			}
		}

		slots = append(slots, dto.AvailableSlotResponse{
			TimeSlot:        slot,
			AvailableTables: available,
		})
	}

	_ = date // parsed for validation only
	return dto.AvailableSlotsResponse{Date: dateStr, Slots: slots}, nil
}

func (s *ReservationService) Book(req dto.BookReservationRequest) (dto.ReservationResponse, error) {
	date, err := parseDate(req.ReservationDate)
	if err != nil {
		return dto.ReservationResponse{}, err
	}

	if !timeslot.Valid(req.TimeSlot) {
		return dto.ReservationResponse{}, fmt.Errorf("%w: invalid time slot", apperror.ErrInvalidInput)
	}

	table, err := s.resolveTable(req, date.Format("2006-01-02"), req.TimeSlot)
	if err != nil {
		return dto.ReservationResponse{}, err
	}

	if req.GuestCount > table.Capacity {
		return dto.ReservationResponse{}, apperror.ErrCapacityExceeded
	}

	existing, err := s.reservationRepo.FindConfirmedByTableDateSlot(table.ID, req.ReservationDate, req.TimeSlot)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.ReservationResponse{}, err
	}
	if existing != nil {
		return dto.ReservationResponse{}, apperror.ErrTableAlreadyBooked
	}

	reservation := &domain.Reservation{
		TableID:         table.ID,
		CustomerName:    req.CustomerName,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		GuestCount:      req.GuestCount,
		ReservationDate: date,
		TimeSlot:        req.TimeSlot,
		SpecialRequests: req.SpecialRequests,
		Status:          domain.ReservationStatusConfirmed,
	}

	if err := s.reservationRepo.Create(reservation); err != nil {
		return dto.ReservationResponse{}, err
	}

	reservation.Table = *table
	return dto.ToReservationResponse(*reservation), nil
}

func (s *ReservationService) GetByID(id uint) (dto.ReservationResponse, error) {
	reservation, err := s.reservationRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ReservationResponse{}, apperror.ErrNotFound
		}
		return dto.ReservationResponse{}, err
	}
	return dto.ToReservationResponse(*reservation), nil
}

// SetNowFunc overrides time source (used in tests).
func (s *ReservationService) SetNowFunc(fn func() time.Time) {
	s.now = fn
}

func (s *ReservationService) Cancel(id uint) (dto.CancelReservationResponse, error) {
	reservation, err := s.reservationRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CancelReservationResponse{}, apperror.ErrNotFound
		}
		return dto.CancelReservationResponse{}, err
	}

	if reservation.Status == domain.ReservationStatusCancelled {
		return dto.CancelReservationResponse{}, apperror.ErrAlreadyCancelled
	}

	slotTime, err := timeslot.DateTime(reservation.ReservationDate, reservation.TimeSlot)
	if err != nil {
		return dto.CancelReservationResponse{}, err
	}

	lead := time.Duration(s.cancelLeadHours) * time.Hour
	if s.now().Add(lead).After(slotTime) {
		return dto.CancelReservationResponse{}, apperror.ErrCancellationDenied
	}

	reservation.Status = domain.ReservationStatusCancelled
	if err := s.reservationRepo.Update(reservation); err != nil {
		return dto.CancelReservationResponse{}, err
	}

	return dto.CancelReservationResponse{ID: reservation.ID, Status: reservation.Status}, nil
}

func (s *ReservationService) resolveTable(req dto.BookReservationRequest, date, slot string) (*domain.Table, error) {
	if req.TableID != nil {
		table, err := s.tableRepo.FindByID(*req.TableID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperror.ErrNotFound
			}
			return nil, err
		}
		return table, nil
	}

	// Smart assignment: smallest table that fits the party.
	tables, err := s.tableRepo.FindAvailableForSlot(date, slot, req.GuestCount)
	if err != nil {
		return nil, err
	}
	if len(tables) == 0 {
		return nil, apperror.ErrNoAvailableTable
	}
	return &tables[0], nil
}

func parseDate(dateStr string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: date must be YYYY-MM-DD", apperror.ErrInvalidInput)
	}
	return date, nil
}

func toSet(ids []uint) map[uint]bool {
	set := make(map[uint]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}
