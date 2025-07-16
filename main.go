package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"strconv"

	"go-book-api/database"
	"go-book-api/models"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()

	router := gin.Default()
	router.POST("/books", createBook)
	router.GET("/books", getBooks)
	router.GET("/books/:id", getBookByID)
	router.PUT("/books/:id", updateBook)
	router.DELETE("/books/:id", deleteBook)
	log.Fatal(router.Run(":8080"))
}

func createBook(c *gin.Context) {
	var book models.Book

	if err := c.ShouldBindJSON(&book); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fieldErrors := make(map[string]string)
			for _, fe := range ve {
				switch fe.Field() {
				case "Title":
					fieldErrors["Title"] = "Title wajib di isi"
				case "Author":
					fieldErrors["Author"] = "Author wajib di isi"
				case "PublicationYear":
					fieldErrors["PublicationYear"] = "PublicationYear wajib di isi"
				default:
					fieldErrors[fe.Field()] = fmt.Sprintf("%s tidak valid", fe.Field())
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "field_validation",
				"message": "Beberapa field wajib di isi",
				"fields":  fieldErrors,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO books (title, author, publication_year) VALUES (?, ?, ?)",
		book.Title, book.Author, book.PublicationYear,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan buku: " + err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendapatkan ID buku baru: " + err.Error()})
		return
	}
	book.ID = int(id)

	c.JSON(http.StatusCreated, book)
}

func getBooks(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, title, author, publication_year FROM books")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil buku: " + err.Error()})
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.PublicationYear); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memindai data buku: " + err.Error()})
			return
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterasi buku: " + err.Error()})
		return
	}

	if len(books) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Tidak ada buku ditemukan.",
			"data":    []models.Book{},
			"total":   0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data buku berhasil diambil",
		"data":    books,
		"total":   len(books),
	})
}

func getBookByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID buku tidak valid"})
		return
	}

	var book models.Book
	row := database.DB.QueryRow("SELECT id, title, author, publication_year FROM books WHERE id = ?", id)
	if err := row.Scan(&book.ID, &book.Title, &book.Author, &book.PublicationYear); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Buku tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil buku: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func updateBook(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID buku tidak valid"})
		return
	}

	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fieldErrors := make(map[string]string)
			for _, fe := range ve {
				switch fe.Field() {
				case "Title":
					fieldErrors["Title"] = "Title wajib di isi"
				case "Author":
					fieldErrors["Author"] = "Author wajib di isi"
				case "PublicationYear":
					fieldErrors["PublicationYear"] = "PublicationYear wajib di isi"
				default:
					fieldErrors[fe.Field()] = "Field ini tidak valid atau kosong"
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "field_validation",
				"message": "Beberapa field wajib di isi atau tidak valid",
				"fields":  fieldErrors,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingBook models.Book

	err = database.DB.QueryRow("SELECT id FROM books WHERE id = ?", id).Scan(&existingBook.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Buku tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa buku: " + err.Error()})
		return
	}

	result, err := database.DB.Exec(
		"UPDATE books SET title = ?, author = ?, publication_year = ? WHERE id = ?",
		book.Title, book.Author, book.PublicationYear, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui buku: " + err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendapatkan jumlah baris terpengaruh: " + err.Error()})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Buku tidak ditemukan untuk diperbarui (mungkin dihapus oleh proses lain)"})
		return
	}

	book.ID = id
	c.JSON(http.StatusOK, book)
}

func deleteBook(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID buku tidak valid"})
		return
	}

	result, err := database.DB.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus buku: " + err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendapatkan affected rows"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Buku tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Buku berhasil dihapus"})
}
