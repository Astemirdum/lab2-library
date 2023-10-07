package model

type ListLibraries struct {
	Paging `json:",inline"`
	Items  []Library
}

type ListBooks struct {
	Paging `json:",inline"`
	Items  []Book
}

type Paging struct {
	Page          int `json:"page"`
	PageSize      int `json:"pageSize"`
	TotalElements int `json:"totalElements"`
}

type Book struct {
	BookUid        string    `json:"bookUid" db:"book_uid"`
	Name           string    `json:"name" db:"name"`
	Author         string    `json:"author" db:"author"`
	Genre          string    `json:"genre" db:"genre"`
	Condition      Condition `json:"condition" db:"condition"`
	AvailableCount int       `json:"availableCount" db:"available_count"`
}

type Condition string

const (
	ConditionExcellent Condition = "EXCELLENT"
	ConditionGood      Condition = "GOOD"
	ConditionBad       Condition = "BAD"
)

type Library struct {
	LibraryUid string `json:"libraryUid" db:"library_uid"`
	Name       string `json:"name" db:"name"`
	Address    string `json:"address" db:"address"`
	City       string `json:"city" db:"city"`
}
