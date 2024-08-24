package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Define the WebMetric model
type WebMetric struct {
	ID             uint      `gorm:"primaryKey"`
	Time           time.Time `gorm:"type:timestamp with time zone"`
	Endpoint       string    `gorm:"size:255"`
	ResponseTimeMs int       `gorm:"not null"`
	StatusCode     int       `gorm:"not null"`
}

// Define the struct for query results
type QueryResult struct {
	Time            time.Time `json:"time"`
	Endpoint        string    `json:"endpoint"`
	ResponseTimeMs  int       `json:"response_time_ms"`
	StatusCode      int       `json:"status_code"`
	CountStatusCode int       `json:"count_status_code"`
}

// Initialize the database connection
func setupDatabase() (*gorm.DB, error) {
	dsn := "host=localhost user=root password=12345 dbname=timescaledb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Auto-migrate the schema
	db.AutoMigrate(&WebMetric{})
	return db, nil
}

// Function to log web metrics
func logMetric(db *gorm.DB, endpoint string, responseTimeMs int, statusCode int) {
	metric := WebMetric{
		Time:           time.Now().UTC(),
		Endpoint:       endpoint,
		ResponseTimeMs: responseTimeMs,
		StatusCode:     statusCode,
	}
	db.Create(&metric)
}

// Function to fetch query results
func getMetrics(db *gorm.DB) ([]QueryResult, error) {
	var results []QueryResult
	query := `
        SELECT time, endpoint, response_time_ms, status_code, COUNT(*) AS count_status_code
        FROM web_metrics
        GROUP BY time, endpoint, response_time_ms, status_code
    `
	if err := db.Raw(query).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func main() {
	router := gin.Default()
	db, err := setupDatabase()
	if err != nil {
		panic("failed to connect to database")
	}

	router.GET("/hello", func(c *gin.Context) {
		start := time.Now()
		// Simulate some processing time
		time.Sleep(1 * time.Second)
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
		elapsed := time.Since(start).Milliseconds()
		logMetric(db, "/hello", int(elapsed), 200)
	})

	router.GET("/metrics", func(c *gin.Context) {
		results, err := getMetrics(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve metrics",
			})
			return
		}
		c.JSON(http.StatusOK, results)
	})

	router.Run(":8080")
}
