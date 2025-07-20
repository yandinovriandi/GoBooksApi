package middlewares

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthRequired adalah middleware yang memastikan pengguna sudah login.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("userID")

		if userID == nil {
			// Pengguna belum login, redirect ke halaman login dengan pesan error
			c.Redirect(http.StatusFound, "/login?error=You need to login first.")
			c.Abort() // Penting: Hentikan pemrosesan request selanjutnya
			return
		}
		// Jika sudah login, lanjutkan ke handler selanjutnya
		c.Next()
	}
}

// AuthorizeRole adalah middleware yang memeriksa peran (role) pengguna.
// roles: daftar peran yang diizinkan (misal: "admin", "editor")
func AuthorizeRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userRole := session.Get("userRole")

		if userRole == nil {
			// Jika tidak ada peran (tidak mungkin terjadi jika AuthRequired() sudah lewat),
			// atau jika sesi somehow rusak, redirect ke login
			c.Redirect(http.StatusFound, "/login?error=Session expired or invalid.")
			c.Abort()
			return
		}

		// Periksa apakah peran pengguna ada di daftar peran yang diizinkan
		isAuthorized := false
		for _, role := range allowedRoles {
			if userRole == role {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			// Pengguna tidak memiliki peran yang diizinkan, kirim error 403 Forbidden
			c.HTML(http.StatusForbidden, "dashboard.html", gin.H{
				"Title":          "Access Denied",
				"WelcomeMessage": "Akses Ditolak!",
				"Username":       session.Get("username"),
				"UserRole":       userRole,
				"Error":          "You do not have sufficient permissions to access this page.",
			})
			c.Abort()
			return
		}
		// Jika pengguna diizinkan, lanjutkan
		c.Next()
	}
}
