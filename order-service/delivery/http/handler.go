package http

import (
	"net/http"

	"github.com/Yessenchik/order-service/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc *usecase.OrderUsecase
}

func NewHandler(uc *usecase.OrderUsecase) *Handler {
	return &Handler{uc: uc}
}

type CreateOrderRequest struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ok, msg, err := h.uc.CreateOrder(c.Request.Context(), req.OrderID, req.Amount, req.CustomerEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": req.OrderID,
		"paid":     ok,
		"message":  msg,
	})
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.uc.UpdateOrderStatus(c.Request.Context(), orderID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"status":   req.Status,
	})
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	order, err := h.uc.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}
