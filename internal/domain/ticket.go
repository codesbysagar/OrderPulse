package domain

import (
	"errors"
	"time"
)

type TicketStatus string

const (
	StatusQueued     TicketStatus = "QUEUED"
	StatusInProgress TicketStatus = "IN_PROGRESS"
	StatusCompleted  TicketStatus = "COMPLETED"
)

type Ticket struct {
	ID            int64
	Title         string
	CreatedBy     string
	Status        TicketStatus
	CreatedAt     time.Time
	SLADeadline   time.Time
	CompletedAt   *time.Time
	BreachAlerted bool
}

func NewTicket(id int64, title, createdBy string, slaDuration time.Duration) (*Ticket, error) {
	if title == "" {
		return nil, errors.New("ticket title cannot be empty")
	}
	if slaDuration <= 0 {
		return nil, errors.New("SLA duration must be positive")
	}

	now := time.Now().UTC()
	return &Ticket{
		ID:          id,
		Title:       title,
		CreatedBy:   createdBy,
		Status:      StatusQueued,
		CreatedAt:   now,
		SLADeadline: now.Add(slaDuration),
	}, nil
}

func (t *Ticket) RemainingTime() time.Duration {
	if t.Status == StatusCompleted {
		return 0
	}
	return time.Until(t.SLADeadline)
}

func (t *Ticket) IsSLABreached() bool {
	if t.Status == StatusCompleted {
		return false
	}
	return time.Now().UTC().After(t.SLADeadline)
}

func (t *Ticket) Start() error {
	if t.Status != StatusQueued {
		return errors.New("only queued tickets can be started")
	}
	t.Status = StatusInProgress
	return nil
}

func (t *Ticket) Complete() error {
	if t.Status == StatusCompleted {
		return errors.New("ticket is already completed")
	}
	now := time.Now().UTC()
	t.Status = StatusCompleted
	t.CompletedAt = &now
	return nil
}
