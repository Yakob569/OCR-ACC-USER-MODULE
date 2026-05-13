package main
import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
    "os"
)
func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("Connect error: %v\n", err)
		return
	}
	imageID, _ := uuid.Parse("20383200-6dde-4c96-aa3a-40adce7a67a1")
	var id uuid.UUID
	err = conn.QueryRow(context.Background(), "SELECT id FROM receipt_extractions WHERE receipt_image_id = $1", imageID).Scan(&id)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
	} else {
		fmt.Printf("Extraction found with ID: %v\n", id)
	}
}
