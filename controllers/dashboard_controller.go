package controllers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

// ShowDashboardPage menampilkan halaman dashboard
func ShowDashboardPage(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username").(string) // Ambil username dari sesi
	userRole := session.Get("userRole").(string) // Ambil role dari sesi

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Title":          "Dashboard Admin Panel",
		"WelcomeMessage": "Selamat datang, " + username + "!",
		"Username":       username,
		"UserRole":       userRole,
	})
}
