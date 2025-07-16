package models

//type Book struct {
//	ID              int    `json:"id"`
//	Title           string `json:"title" binding:"required"`
//	Author          string `json:"author" binding:"required"`
//	PublicationYear int    `json:"publication_year,omitempty"` //ini artinya boleh tidak di isi
//}

type Book struct {
	ID              int    `json:"id"`
	Title           string `json:"title" binding:"required"`
	Author          string `json:"author" binding:"required"`
	PublicationYear int    `json:"publication_year" binding:"required"`
}
