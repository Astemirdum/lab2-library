package model

type Rating struct {
	Stars int `json:"stars" db:"stars"`
}
