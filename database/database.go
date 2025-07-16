package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
	dsn := "root:@tcp(127.0.0.1:3306)/book_management?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Koneksi database tidak valid: %v", err)
	}

	fmt.Println("Koneksi database MySQL berhasil!")
}
