package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func main() {
	// Path relatif atau absolut yang fleksibel
	// Pastikan folder data/ ada
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Gagal membuat direktori data: %v", err)
	}

	dbPath := filepath.Join(dataDir, "app.db")

	// Jalankan / buka database SQLite menggunakan driver modernc.org/sqlite
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Gagal membuka database sqlite: %v", err)
	}
	defer db.Close()

	// Cari file schema.sql dan seed.sql
	// Cek path baik dijalankan dari backend/ maupun backend/cmd/initdb/
	schemaPath := findFile("database/schema.sql", "../../database/schema.sql", "schema.sql")
	seedPath := findFile("database/seed.sql", "../../database/seed.sql", "seed.sql")

	if schemaPath != "" {
		executeSQLFile(db, schemaPath)
		fmt.Printf("Berhasil mengeksekusi: %s\n", schemaPath)
	} else {
		log.Println("Peringatan: schema.sql tidak ditemukan")
	}

	if seedPath != "" {
		executeSQLFile(db, seedPath)
		fmt.Printf("Berhasil mengeksekusi: %s\n", seedPath)
	} else {
		log.Println("Peringatan: seed.sql tidak ditemukan")
	}

	// Cetak jumlah baris di tabel servis
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM gallery").Scan(&count)
	if err != nil {
		// Jika tabel servis tidak ada atau error
		fmt.Printf("Error menghitung tabel gallery: %v\n", err)
	} else {
		fmt.Printf("Jumlah baris di tabel gallery: %d\n", count)
	}
}

func findFile(candidates ...string) string {
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func executeSQLFile(db *sql.DB, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Gagal membaca file %s: %v", filePath, err)
	}

	if len(content) == 0 {
		return
	}

	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("Gagal mengeksekusi sql dari %s: %v", filePath, err)
	}
}
