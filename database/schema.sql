-- Schema database untuk portal Desa Makmur
-- Berdasarkan Blueprint Bagian 3 (Data)

-- 3.1 Model Data Galeri Kegiatan (Gallery / Activity Media)
CREATE TABLE IF NOT EXISTS gallery (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(150) NOT NULL,
    caption TEXT,
    category VARCHAR(50) NOT NULL,
    uploader_name VARCHAR(100) NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    thumbnail_url VARCHAR(255),
    file_size INT,
    mime_type VARCHAR(50),
    activity_date DATE,
    status VARCHAR(20) DEFAULT 'published',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);