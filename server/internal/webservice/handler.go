package webservice

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetServices(c *gin.Context) {
	services, err := h.repo.GetAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, services)
}

func (h *Handler) CreateService(c *gin.Context) {
	var service WebService

	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := h.repo.Create(c, &service)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, service)
}

func (h *Handler) GetServicesBySiteID(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid site id"})
		return
	}

	services, err := h.repo.GetBySiteID(c, siteID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, services)
}
