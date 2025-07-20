package main

import (
	"database/sql" // Tetap dibutuhkan untuk CRUD buku lama
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go-book-api/controllers" // Import controllers baru
	"go-book-api/database"
	"go-book-api/middlewares" // Import middlewares baru
	"go-book-api/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	database.InitDB() // Menginisialisasi koneksi DB (GORM dan SQL murni)

	router := gin.Default()

	// Konfigurasi sesi (gunakan secret key yang kuat di produksi!)
	// Secret key harus panjang dan acak
	store := cookie.NewStore([]byte("this-is-a-very-secret-key-for-your-gin-session-app-please-change-it-to-a-random-one-in-production-!!!!!!"))
	router.Use(sessions.Sessions("mysession", store))

	// Muat template HTML dari folder views dan sub-direktorinya
	router.LoadHTMLGlob("views/**/*.html")

	// Servis file statis (CSS, JS) dari folder public
	router.Static("/public", "./public")

	// --- Rute API Buku (yang sudah ada) ---
	// Rute-rute ini tetap berfungsi sebagai API JSON
	apiRoutes := router.Group("/api")
	{
		apiRoutes.POST("/books", createBook)
		apiRoutes.GET("/books", getBooks)
		apiRoutes.GET("/books/:id", getBookByID)
		apiRoutes.PUT("/books/:id", updateBook)
		apiRoutes.DELETE("/books/:id", deleteBook)
	}

	// --- Rute Web (untuk Dashboard & Autentikasi) ---

	// Grup rute untuk login dan register (tidak memerlukan otentikasi)
	authRoutes := router.Group("/")
	{
		authRoutes.GET("/login", controllers.ShowLoginPage)
		authRoutes.POST("/login", controllers.Login)
		authRoutes.POST("/register", controllers.Register) // Endpoint register (untuk setup awal admin/user)
	}

	// Grup rute yang memerlukan otentikasi (semua user yang login)
	authenticatedRoutes := router.Group("/")
	authenticatedRoutes.Use(middlewares.AuthRequired()) // Semua rute di grup ini memerlukan login
	{
		authenticatedRoutes.GET("/dashboard", controllers.ShowDashboardPage)
		authenticatedRoutes.GET("/logout", controllers.Logout)

		// Contoh rute khusus untuk user biasa (misal halaman profil mereka sendiri)
		authenticatedRoutes.GET("/profile", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Title":          "User Profile",
				"WelcomeMessage": "Ini halaman profil Anda.",
				"Username":       sessions.Default(c).Get("username"),
				"UserRole":       sessions.Default(c).Get("userRole"),
				"ContentTitle":   "Profil Pengguna",
				"ContentBody":    "Di sini Anda dapat mengelola informasi profil Anda.",
			})
		})

		// --- Rute untuk Manajemen Buku di Dashboard (akan kita buat di langkah selanjutnya) ---
		// authenticatedRoutes.GET("/dashboard/books", controllers.ShowBookManagementPage)
		// authenticatedRoutes.GET("/dashboard/books/new", controllers.ShowNewBookForm)
		// authenticatedRoutes.POST("/dashboard/books/new", controllers.CreateBookFromForm)
		// ... dll.
	}

	// Grup rute khusus ADMIN (memerlukan login DAN peran admin)
	adminRoutes := router.Group("/admin")
	adminRoutes.Use(middlewares.AuthRequired())         // Memastikan sudah login
	adminRoutes.Use(middlewares.AuthorizeRole("admin")) // Memastikan role adalah 'admin'
	{
		adminRoutes.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Title":          "Admin Dashboard",
				"WelcomeMessage": "Selamat datang, Admin " + sessions.Default(c).Get("username").(string) + "!",
				"Username":       sessions.Default(c).Get("username"),
				"UserRole":       sessions.Default(c).Get("userRole"),
				"ContentTitle":   "Area Khusus Admin",
				"ContentBody":    "Ini adalah halaman yang hanya bisa diakses oleh administrator.",
			})
		})
		adminRoutes.GET("/users", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Title":          "Manajemen Pengguna",
				"WelcomeMessage": "Selamat datang, Admin " + sessions.Default(c).Get("username").(string) + "!",
				"Username":       sessions.Default(c).Get("username"),
				"UserRole":       sessions.Default(c).Get("userRole"),
				"ContentTitle":   "Daftar Pengguna Sistem",
				"ContentBody":    "Di sini Anda bisa melihat dan mengelola daftar pengguna.",
			})
		})
		// Tambahkan rute admin lainnya di sini, misalnya untuk CRUD buku dari sisi admin
		// adminRoutes.GET("/books-admin", controllers.ShowAdminBookManagementPage)
	}

	// Rute default redirect ke login
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/login")
	})

	log.Fatal(router.Run(":8080"))
}

// --- Fungsi-fungsi CRUD Buku yang Sudah Ada (tetap di sini untuk API) ---
// Perhatikan: fungsi-fungsi ini masih menggunakan database.DB_RAW
// Jika kamu ingin mengubahnya ke GORM, kamu perlu memodifikasi implementasinya
// dan menggunakan database.DB_GORM.

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

	result, err := database.DB_RAW.Exec( // Menggunakan DB_RAW
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
	rows, err := database.DB_RAW.Query("SELECT id, title, author, publication_year FROM books") // Menggunakan DB_RAW
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
	row := database.DB_RAW.QueryRow("SELECT id, title, author, publication_year FROM books WHERE id = ?", id) // Menggunakan DB_RAW
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

	err = database.DB_RAW.QueryRow("SELECT id FROM books WHERE id = ?", id).Scan(&existingBook.ID) // Menggunakan DB_RAW
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Buku tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa buku: " + err.Error()})
		return
	}

	result, err := database.DB_RAW.Exec( // Menggunakan DB_RAW
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

	result, err := database.DB_RAW.Exec("DELETE FROM books WHERE id = ?", id) // Menggunakan DB_RAW
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
