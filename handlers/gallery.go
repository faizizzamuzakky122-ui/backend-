package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GalleryHandler struct {
	DB *sql.DB
}

func NewGalleryHandler(db *sql.DB) *GalleryHandler {
	return &GalleryHandler{DB: db}
}

// 4.1 GET /api/v1/gallery
func (h *GalleryHandler) GetGalleries(c *gin.Context) {
	category := c.Query("category")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "12")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 12
	}

	offset := (page - 1) * limit

	var totalRecords int
	countQuery := "SELECT COUNT(*) FROM gallery WHERE status = 'published'"
	var countArgs []interface{}

	if category != "" {
		countQuery += " AND category = ?"
		countArgs = append(countArgs, category)
	}

	err = h.DB.QueryRow(countQuery, countArgs...).Scan(&totalRecords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menghitung total data galeri",
		})
		return
	}

	selectQuery := `
		SELECT id, title, caption, category, uploader_name, image_url, 
		       thumbnail_url, file_size, mime_type, activity_date, status, created_at, updated_at
		FROM gallery 
		WHERE status = 'published'
	`
	var selectArgs []interface{}

	if category != "" {
		selectQuery += " AND category = ?"
		selectArgs = append(selectArgs, category)
	}

	selectQuery += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	selectArgs = append(selectArgs, limit, offset)

	rows, err := h.DB.Query(selectQuery, selectArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data galeri",
		})
		return
	}
	defer rows.Close()

	var galleries []models.Gallery
	for rows.Next() {
		var g models.Gallery
		var caption, thumbnailURL, mimeType, activityDate, updatedAt sql.NullString
		var fileSize sql.NullInt64

		err := rows.Scan(
			&g.ID,
			&g.Title,
			&caption,
			&g.Category,
			&g.UploaderName,
			&g.ImageURL,
			&thumbnailURL,
			&fileSize,
			&mimeType,
			&activityDate,
			&g.Status,
			&g.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Gagal membaca baris data galeri",
			})
			return
		}

		if caption.Valid {
			g.Caption = &caption.String
		}
		if thumbnailURL.Valid {
			g.ThumbnailURL = &thumbnailURL.String
		}
		if fileSize.Valid {
			fs := int(fileSize.Int64)
			g.FileSize = &fs
		}
		if mimeType.Valid {
			g.MimeType = &mimeType.String
		}
		if activityDate.Valid {
			g.ActivityDate = &activityDate.String
		}
		if updatedAt.Valid {
			g.UpdatedAt = &updatedAt.String
		}

		galleries = append(galleries, g)
	}

	if galleries == nil {
		galleries = []models.Gallery{}
	}

	totalPages := 0
	if totalRecords > 0 {
		totalPages = int(math.Ceil(float64(totalRecords) / float64(limit)))
	}

	c.JSON(http.StatusOK, models.GalleryListResponse{
		Status: "success",
		Data:   galleries,
		Pagination: models.Pagination{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalRecords: totalRecords,
		},
	})
}

// 4.2 POST /api/v1/gallery/upload
func (h *GalleryHandler) UploadGallery(c *gin.Context) {
	title := strings.TrimSpace(c.PostForm("title"))
	category := strings.TrimSpace(c.PostForm("category"))
	uploaderName := strings.TrimSpace(c.PostForm("uploader_name"))
	caption := strings.TrimSpace(c.PostForm("caption"))
	activityDate := strings.TrimSpace(c.PostForm("activity_date"))

	if title == "" || category == "" || uploaderName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "title, category, dan uploader_name wajib diisi",
		})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Berkas foto (image) wajib diunggah",
		})
		return
	}

	// Validasi ukuran berkas (maks 5 MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Ukuran berkas melebihi batas maksimal 5 MB",
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format berkas tidak didukung. Gunakan JPG, PNG, atau WebP",
		})
		return
	}

	// Buat folder simpanan jika belum ada
	uploadDir := filepath.Join("uploads", "gallery")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menyiapkan direktori penyimpanan berkas",
		})
		return
	}

	id := uuid.New().String()
	timestamp := time.Now().Unix()
	newFilename := fmt.Sprintf("galeri_%s_%d%s", id[:8], timestamp, ext)
	savePath := filepath.Join(uploadDir, newFilename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menyimpan berkas gambar",
		})
		return
	}

	imageURL := fmt.Sprintf("/uploads/gallery/%s", newFilename)
	now := time.Now().UTC().Format(time.RFC3339)
	fileSize := int(file.Size)
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/" + strings.TrimPrefix(ext, ".")
	}

	var actDateParam interface{}
	var actDateRes *string
	if activityDate != "" {
		actDateParam = activityDate
		actDateRes = &activityDate
	} else {
		actDateParam = nil
	}

	var captionParam interface{}
	if caption != "" {
		captionParam = caption
	} else {
		captionParam = nil
	}

	insertQuery := `
		INSERT INTO gallery (
			id, title, caption, category, uploader_name, 
			image_url, thumbnail_url, file_size, mime_type, 
			activity_date, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'published', ?, ?)
	`

	_, err = h.DB.Exec(
		insertQuery,
		id,
		title,
		captionParam,
		category,
		uploaderName,
		imageURL,
		imageURL, // thumbnail url default
		fileSize,
		mimeType,
		actDateParam,
		now,
		now,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Gagal menyimpan data galeri ke database: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, models.GalleryUploadResponse{
		Status:  "success",
		Message: "Foto kegiatan berhasil diunggah dan ditambahkan ke galeri desa",
		Data: models.GalleryCreated{
			ID:           id,
			Title:        title,
			Category:     category,
			UploaderName: uploaderName,
			ImageURL:     imageURL,
			ActivityDate: actDateRes,
			CreatedAt:    now,
		},
	})
}

// 4.3 GET /api/v1/gallery/:id
func (h *GalleryHandler) GetGalleryByID(c *gin.Context) {
	id := c.Param("id")

	query := `
		SELECT id, title, caption, category, uploader_name, image_url,
		       thumbnail_url, file_size, mime_type, activity_date, status, created_at, updated_at
		FROM gallery
		WHERE id = ?
	`

	var g models.Gallery
	var caption, thumbnailURL, mimeType, activityDate, updatedAt sql.NullString
	var fileSize sql.NullInt64

	err := h.DB.QueryRow(query, id).Scan(
		&g.ID,
		&g.Title,
		&caption,
		&g.Category,
		&g.UploaderName,
		&g.ImageURL,
		&thumbnailURL,
		&fileSize,
		&mimeType,
		&activityDate,
		&g.Status,
		&g.CreatedAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Data galeri tidak ditemukan",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil rincian data galeri",
		})
		return
	}

	if caption.Valid {
		g.Caption = &caption.String
	}
	if thumbnailURL.Valid {
		g.ThumbnailURL = &thumbnailURL.String
	}
	if fileSize.Valid {
		fs := int(fileSize.Int64)
		g.FileSize = &fs
	}
	if mimeType.Valid {
		g.MimeType = &mimeType.String
	}
	if activityDate.Valid {
		g.ActivityDate = &activityDate.String
	}
	if updatedAt.Valid {
		g.UpdatedAt = &updatedAt.String
	}

	c.JSON(http.StatusOK, models.GalleryDetailResponse{
		Status: "success",
		Data:   g,
	})
}
