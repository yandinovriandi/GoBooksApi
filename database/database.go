package database

import (
	"database/sql"
	"fmt"
	"go-book-api/models" // Import models kamu
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql" // Driver MySQL
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB_RAW *sql.DB   // Untuk operasi buku yang menggunakan database/sql
var DB_GORM *gorm.DB // Untuk operasi user dan GORM lainnya

// InitDB menginisialisasi koneksi database untuk sql.DB dan GORM
func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME") // Nama database untuk user dan GORM

	// Koneksi untuk GORM (untuk User model)
	dsnGorm := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
	databaseGorm, err := gorm.Open(mysql.Open(dsnGorm), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database (GORM): %v", err)
	}
	DB_GORM = databaseGorm
	fmt.Println("Koneksi database MySQL (GORM) berhasil!")

	// Migrasi database untuk User model
	err = DB_GORM.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Gagal auto migrate User model: %v", err)
	}
	log.Println("Database migration completed for User model!")

	// Koneksi untuk database/sql (untuk Book model, jika masih ingin pakai ini)
	// Kita akan menggunakan nama database yang sama untuk kesederhanaan
	dsnRaw := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
	var errRaw error
	DB_RAW, errRaw = sql.Open("mysql", dsnRaw)
	if errRaw != nil {
		log.Fatalf("Gagal terhubung ke database (SQL murni): %v", errRaw)
	}
	if errRaw = DB_RAW.Ping(); errRaw != nil {
		log.Fatalf("Koneksi database (SQL murni) tidak valid: %v", errRaw)
	}
	fmt.Println("Koneksi database MySQL (SQL murni) berhasil!")
}
