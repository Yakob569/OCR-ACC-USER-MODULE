package config

import "testing"

func TestLoadConfigRejectsDefaultJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "change-me-at-all-costs")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected LoadConfig to reject default JWT secret")
	}
}

func TestLoadConfigLoadsOCRAndMinIOSettings(t *testing.T) {
	t.Setenv("JWT_SECRET", "super-secret-value")
	t.Setenv("OCR_ENGINE_BASE_URL", "https://ocr.example.com")
	t.Setenv("OCR_ENGINE_TIMEOUT_SECONDS", "45")
	t.Setenv("OCR_ENGINE_MAX_CONCURRENCY", "6")
	t.Setenv("OCR_GROUP_MAX_FILES", "50")
	t.Setenv("OCR_MAX_FILE_SIZE_MB", "15")
	t.Setenv("MINIO_ACCESS_KEY_ID", "minio-access")
	t.Setenv("MINIO_SECRET_ACCESS_KEY", "minio-secret")
	t.Setenv("MINIO_BUCKET_NAME", "receipts")
	t.Setenv("MINIO_END_POINT", "minio.example.com")
	t.Setenv("MINIO_CLAMAV_URL", "https://clamav.example.com")
	t.Setenv("MINIO_USE_SSL", "false")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}

	if cfg.OCREngine.BaseURL != "https://ocr.example.com" {
		t.Fatalf("unexpected OCR engine base URL: %s", cfg.OCREngine.BaseURL)
	}
	if cfg.OCREngine.TimeoutSeconds != 45 {
		t.Fatalf("unexpected OCR engine timeout: %d", cfg.OCREngine.TimeoutSeconds)
	}
	if cfg.OCREngine.MaxConcurrency != 6 {
		t.Fatalf("unexpected OCR engine concurrency: %d", cfg.OCREngine.MaxConcurrency)
	}
	if cfg.OCRGroupMaxFiles != 50 {
		t.Fatalf("unexpected OCR group max files: %d", cfg.OCRGroupMaxFiles)
	}
	if cfg.OCRMaxFileSizeMB != 15 {
		t.Fatalf("unexpected OCR max file size: %d", cfg.OCRMaxFileSizeMB)
	}
	if cfg.MinIO.AccessKeyID != "minio-access" {
		t.Fatalf("unexpected MinIO access key ID: %s", cfg.MinIO.AccessKeyID)
	}
	if cfg.MinIO.SecretAccessKey != "minio-secret" {
		t.Fatalf("unexpected MinIO secret access key: %s", cfg.MinIO.SecretAccessKey)
	}
	if cfg.MinIO.BucketName != "receipts" {
		t.Fatalf("unexpected MinIO bucket name: %s", cfg.MinIO.BucketName)
	}
	if cfg.MinIO.Endpoint != "minio.example.com" {
		t.Fatalf("unexpected MinIO endpoint: %s", cfg.MinIO.Endpoint)
	}
	if cfg.MinIO.CLAMAVURL != "https://clamav.example.com" {
		t.Fatalf("unexpected MinIO CLAMAV URL: %s", cfg.MinIO.CLAMAVURL)
	}
	if cfg.MinIO.UseSSL {
		t.Fatal("expected MinIO USE_SSL to be false")
	}
}

func TestLoadConfigRejectsInvalidOCRTimeout(t *testing.T) {
	t.Setenv("JWT_SECRET", "super-secret-value")
	t.Setenv("OCR_ENGINE_TIMEOUT_SECONDS", "invalid")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected invalid OCR timeout to fail")
	}
}
