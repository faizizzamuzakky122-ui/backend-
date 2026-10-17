package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"backend/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// Middleware CORS khusus mengizinkan dev server Vue (Vite port 5173 atau dinamis localhost Vite)
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Izinkan dev server Vue (misalnya http://localhost:5173, http://127.0.0.1:5173)
		if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	// Muat file .env jika ada
	_ = godotenv.Load(".env")

	port := os.Getenv("PORT")
	if strings.TrimSpace(port) == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if strings.TrimSpace(dbPath) == "" {
		dbPath = "data/app.db"
	}
	dbPath = strings.TrimSpace(dbPath)

	// Pastikan folder database ada
	dbDir := filepath.Dir(dbPath)
	if dbDir != "" && dbDir != "." {
		_ = os.MkdirAll(dbDir, 0755)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database sqlite: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database tidak merespon: %v", err)
	}

	r := gin.Default()

	// Terapkan CORS middleware
	r.Use(corsMiddleware())

	// Layani file statis yang diunggah
	r.Static("/uploads", "./uploads")

	galleryHandler := handlers.NewGalleryHandler(db)

	// Routing REST API sesuai Blueprint
	apiV1 := r.Group("/api/v1")
	{
		galleryGroup := apiV1.Group("/gallery")
		{
			galleryGroup.GET("", galleryHandler.GetGalleries)
			galleryGroup.POST("/upload", galleryHandler.UploadGallery)
			galleryGroup.GET("/:id", galleryHandler.GetGalleryByID)
		}
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server berjalan di port %s, database: %s", port, dbPath)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Gagal menjalankan server Gin: %v", err)
	}
}
