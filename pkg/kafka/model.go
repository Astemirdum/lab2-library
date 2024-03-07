package kafka

import "time"

type EventStats struct {
	Timestamp     time.Time `json:"timestamp"`
	UserName      string    `json:"username"`
	ReservationID string    `json:"reservation_uid"`
	BookID        string    `json:"book_uid"`
	LibraryID     string    `json:"library_uid"`
	Rating        int       `json:"rating"`
	EventType     string    `json:"event_type"`
	Simplex       Simplex   `json:"simplex"`
}

type Simplex string

const (
	SimplexUnknown Simplex = "UNKNOWN"
	SimplexUp      Simplex = "UP"
	SimplexDown    Simplex = "DOWN"
)
