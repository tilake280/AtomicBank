package api

import (
	"atomicbank/events"
	"atomicbank/ledger"
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransferRequest struct {
	ToAccountID string          `json:"to_account_id" binding:"required,uuid"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		// Simulated Auth: Hardcoded to map directly to our seeded Account 1
		senderAccountID, _ := uuid.Parse("11111111-1111-1111-1111-111111111111")

		c.Set("sender_account_id", senderAccountID)
		c.Next()
	}
}

func TransferHandler(db *sql.DB, bus *events.EventBus) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TransferRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
			return
		}

		senderCtxValue, exists := c.Get("sender_account_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth context missing"})
			return
		}
		fromAccountID := senderCtxValue.(uuid.UUID)

		toAccountID, err := uuid.Parse(req.ToAccountID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid destination account ID"})
			return
		}

		err = ledger.TransferFunds(c.Request.Context(), db, bus, fromAccountID, toAccountID, req.Amount)
		if err != nil {
			if err == ledger.ErrInsufficientFunds {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "insufficient funds"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed to process"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "transfer completed successfully",
		})
	}
}

func BalanceHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		senderCtxValue, exists := c.Get("sender_account_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth context missing"})
			return
		}
		accountID := senderCtxValue.(uuid.UUID)

		var balance decimal.Decimal
		err := db.QueryRowContext(c.Request.Context(), `
			SELECT COALESCE(SUM(amount), 0) 
			FROM ledger_entries 
			WHERE account_id = $1`, accountID).Scan(&balance)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch balance"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"account_id": accountID,
			"balance":    balance,
		})
	}
}
