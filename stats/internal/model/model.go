package model

import "time"

type Stats struct {
	UserName    string    `json:"username" db:"username"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	Rating      int       `json:"rating" db:"rating"`
	CountReserv int       `json:"cnt_reserv" db:"cnt_reserv"`
	CountBooks  int       `json:"cnt_books" db:"cnt_books"`
	CountLibs   int       `json:"cnt_libs" db:"cnt_libs"`
}

type StatsInfo struct {
	Data []Stats `json:"data"`
}
