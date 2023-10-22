package model

import (
	"fmt"
	"strings"
	"time"
)

type CreateReservationResponse struct {
	ReservationUid string  `json:"reservationUid"`
	Status         string  `json:"status"`
	StartDate      Date2   `json:"startDate"`
	TillDate       Date2   `json:"tillDate"`
	Library        Library `json:"library"`
	Book           Book    `json:"book"`
	Rating         Rating  `json:"rating"`
}

type GetReservationResponse struct {
	Reservation `json:",inline"`
	Library     Library `json:"library"`
	Book        Book    `json:"book"`
}

type CreateReservationRequest struct {
	BookUid    string `json:"bookUid" validate:"required"`
	LibraryUid string `json:"libraryUid" validate:"required"`
	TillDate   Date   `json:"tillDate" validate:"required"`
	UserName   string `json:"-" validate:"required"`
	Stars      int    `json:"rating"`
}

type Date struct {
	time.Time `json:",inline"`
}

func (d *Date) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	date, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}
	d.Time = date
	return
}

type Date2 struct {
	time.Time `json:",inline"`
}

func (d Date2) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Local().Format(time.DateOnly))), nil
}

type Library struct {
	LibraryUid string `json:"libraryUid"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
}

type Book struct {
	BookUid string `json:"bookUid"`
	Name    string `json:"name"`
	Author  string `json:"author"`
	Genre   string `json:"genre"`
}

type GetBook struct {
	ID        int `json:"id"`
	Book      `json:",inline"`
	Condition string `json:"condition"`
}

type GetLibrary struct {
	ID      int `json:"id"`
	Library `json:",inline"`
}

type Reservation struct {
	ReservationUid string    `json:"reservationUid"`
	Status         string    `json:"status"`
	StartDate      time.Time `json:"startDate"`
	TillDate       time.Time `json:"tillDate"`
}

type GetReservation struct {
	ReservationUid string    `json:"reservationUid"`
	BookUid        string    `json:"bookUid"`
	LibraryUid     string    `json:"libraryUid"`
	Status         string    `json:"status"`
	StartDate      time.Time `json:"startDate"`
	TillDate       time.Time `json:"tillDate"`
}

type Rating struct {
	Stars int `json:"stars"`
}

type RatingMsg struct {
	Name  string `json:"name"`
	Stars int    `json:"stars"`
}

type ReservationReturnRequest struct {
	Condition string `json:"condition" validate:"required,oneof=EXCELLENT GOOD BAD"`
	Date      Date   `json:"date" validate:"required"`
}

type ReservationReturnResponse struct {
	BookUid    string `json:"bookUid" db:"book_uid"`
	LibraryUid string `json:"libraryUid" db:"library_uid"`
}

type AvailableCountRequest struct {
	LibraryID int  `json:"libraryID"`
	BookID    int  `json:"bookID"`
	IsReturn  bool `json:"isReturn"`
}
