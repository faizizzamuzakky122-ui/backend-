package models

type Gallery struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Caption      *string `json:"caption,omitempty"`
	Category     string  `json:"category"`
	UploaderName string  `json:"uploader_name"`
	ImageURL     string  `json:"image_url"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`
	FileSize     *int    `json:"file_size,omitempty"`
	MimeType     *string `json:"mime_type,omitempty"`
	ActivityDate *string `json:"activity_date,omitempty"`
	Status       string  `json:"status,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    *string `json:"updated_at,omitempty"`
}

type Pagination struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

type GalleryListResponse struct {
	Status     string      `json:"status"`
	Data       []Gallery   `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type GalleryDetailResponse struct {
	Status string  `json:"status"`
	Data   Gallery `json:"data"`
}

type GalleryUploadResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    GalleryCreated `json:"data"`
}

type GalleryCreated struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Category     string  `json:"category"`
	UploaderName string  `json:"uploader_name"`
	ImageURL     string  `json:"image_url"`
	ActivityDate *string `json:"activity_date,omitempty"`
	CreatedAt    string  `json:"created_at"`
}
