package controllers

import (
	"go-book-api/database" // Ubah import ini ke database yang baru
	"go-book-api/models"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Struktur untuk input login
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"` // Hanya untuk register
}

// ShowLoginPage menampilkan halaman login
func ShowLoginPage(c *gin.Context) {
	// Cek apakah user sudah login, jika ya, redirect ke dashboard
	session := sessions.Default(c)
	if session.Get("userID") != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	errorMessage := c.Query("error")
	c.HTML(http.StatusOK, "login.html", gin.H{"Error": errorMessage})
}

// Login menangani proses autentikasi pengguna
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var user models.User
	// Gunakan DB_GORM untuk query user
	if err := database.DB_GORM.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Set session setelah login berhasil
	session := sessions.Default(c)
	session.Set("userID", user.ID)
	session.Set("username", user.Username) // Simpan username ke sesi
	session.Set("userRole", user.Role)     // Simpan role ke sesi
	if err := session.Save(); err != nil { // Penting: cek error saat menyimpan sesi
		log.Printf("Failed to save session for user %s: %v", user.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "redirect": "/dashboard"})
}

// Logout menghapus sesi pengguna
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()                        // Hapus semua data sesi
	if err := session.Save(); err != nil { // Penting: cek error saat menyimpan sesi
		log.Printf("Failed to clear session: %v", err)
	}
	c.Redirect(http.StatusFound, "/login?error=You have been logged out.")
}

// Register menangani pendaftaran user baru (hanya untuk testing/initial setup)
func Register(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Default role adalah 'user'. Hanya izinkan 'admin' jika memang diinputkan.
	role := "user"
	if input.Role == "admin" {
		role = "admin"
	}

	user := models.User{Username: input.Username, Password: string(hashedPassword), Role: role}
	// Gunakan DB_GORM untuk membuat user
	if err := database.DB_GORM.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user. Perhaps username already exists?"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": gin.H{"username": user.Username, "role": user.Role}})
}
