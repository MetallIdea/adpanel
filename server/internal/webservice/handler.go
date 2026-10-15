package webservice

import (
	"database/sql"
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

func (h *Handler) GetServiceByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid service id"})
		return
	}

	service, err := h.repo.GetServiceByID(c, id)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "service not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, service)
}

func (h *Handler) CreateService(c *gin.Context) {
	var service WebService

	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	if service.Name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}

	if service.Port <= 0 {
		c.JSON(400, gin.H{"error": "port must be a positive number"})
		return
	}

	err := h.repo.Create(c, &service)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, service)
}

func (h *Handler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid service id"})
		return
	}

	var service WebService
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	service.ID = id

	if service.Name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}

	if service.Port <= 0 {
		c.JSON(400, gin.H{"error": "port must be a positive number"})
		return
	}

	// Check if service exists
	_, err = h.repo.GetServiceByID(c, id)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "service not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = h.repo.Update(c, &service)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, service)
}

func (h *Handler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid service id"})
		return
	}

	// Check if service exists
	_, err = h.repo.GetServiceByID(c, id)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "service not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = h.repo.Delete(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
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
