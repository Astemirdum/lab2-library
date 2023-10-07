package model

import (
	"time"
)

type CreateReservationRequest struct {
	BookUid    string `json:"bookUid" validate:"required"`
	LibraryUid string `json:"libraryUid" validate:"required"`
	TillDate   Date   `json:"tillDate" validate:"required"`
	UserName   string `validate:"required"`
}

type Date struct {
	time.Time `json:",inline"`
}

type Status string

const (
	StatusRented   Status = "RENTED"
	StatusReturned Status = "RETURNED"
	StatusExpired  Status = "EXPIRED"
)

type Reservation struct {
	ID             int       `json:"-" db:"id"`
	ReservationUid string    `json:"reservationUid" db:"reservation_uid"`
	Username       string    `json:"username" db:"username"`
	BookUid        string    `json:"bookUid" db:"book_uid"`
	LibraryUid     string    `json:"libraryUid" db:"library_uid"`
	Status         Status    `json:"status" db:"status"`
	StartDate      time.Time `json:"startDate" db:"start_date"`
	TillDate       time.Time `json:"tillDate" db:"till_date"`
}

type ReservationReturnRequest struct {
	Condition string `json:"condition" validate:"required,oneof=EXCELLENT GOOD BAD"`
	Date      Date   `json:"date" validate:"required"`
}
