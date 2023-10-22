package model

type ListLibraries struct {
	Paging `json:",inline"`
	Items  []Library `json:"items"`
}

type ListBooks struct {
	Paging `json:",inline"`
	Items  []Book `json:"items"`
}

type Paging struct {
	Page          int `json:"page"`
	PageSize      int `json:"pageSize"`
	TotalElements int `json:"totalElements"`
}

type Book struct {
	ID             int       `json:"id" db:"id"`
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
	ID         int    `json:"id" db:"id"`
	LibraryUid string `json:"libraryUid" db:"library_uid"`
	Name       string `json:"name" db:"name"`
	Address    string `json:"address" db:"address"`
	City       string `json:"city" db:"city"`
}

type AvailableCountRequest struct {
	LibraryID int  `json:"libraryID"`
	BookID    int  `json:"bookID"`
	IsReturn  bool `json:"isReturn"`
}
