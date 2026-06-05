package main

import (
	"atomicbank/api"
	"atomicbank/events"
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgresql://admin:secret@localhost:5432/atomicbank?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}

	bus := events.NewEventBus()
	bus.StartNotificationWorker()

	router := gin.Default()

	router.StaticFile("/", "./index.html")

	v1 := router.Group("/api/v1")
	{
		secure := v1.Group("/")
		secure.Use(api.AuthMiddleware())
		{
			secure.GET("/balance", api.BalanceHandler(db))
			secure.POST("/transfers", api.TransferHandler(db, bus))
		}
	}

	log.Println("⚛️ AtomicBank API starting on :8080...")
	router.Run(":8080")
}
