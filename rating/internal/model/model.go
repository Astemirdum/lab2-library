package model

type Rating struct {
	Stars int `json:"stars" db:"stars"`
}

type RatingMsg struct {
	Name  string `json:"name"`
	Stars int    `json:"stars"`
}

type CreateRating struct {
	Name  string `json:"name"`
	Stars int    `json:"stars"`
}
