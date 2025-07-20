package models

import "gorm.io/gorm"

// User merepresentasikan model pengguna untuk autentikasi dan otorisasi
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`                // Password yang sudah di-hash
	Role     string `gorm:"default:'user';not null"` // 'admin' atau 'user'
}
