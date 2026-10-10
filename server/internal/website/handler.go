package website

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

func (h *Handler) GetSites(c *gin.Context) {
	sites, err := h.repo.GetAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, sites)
}

func (h *Handler) GetSiteByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid site id"})
		return
	}

	site, err := h.repo.GetByID(c, id)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "site not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, site)
}

func (h *Handler) CreateSite(c *gin.Context) {
	var site Website

	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	site.Name = sanitizeInput(site.Name)
	site.URL = sanitizeInput(site.URL)
	site.Status = sanitizeInput(site.Status)

	if site.Name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}

	if site.URL == "" {
		c.JSON(400, gin.H{"error": "url is required"})
		return
	}

	if !isValidStatus(site.Status) {
		c.JSON(400, gin.H{"error": "invalid status. Allowed values: active, inactive, pending"})
		return
	}

	err := h.repo.Create(c, &site)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, site)
}

func (h *Handler) DeleteSite(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid site id"})
		return
	}

	err = h.repo.Delete(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}

func (h *Handler) UpdateSite(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid site id"})
		return
	}

	var site Website
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	site.ID = id
	site.Name = sanitizeInput(site.Name)
	site.URL = sanitizeInput(site.URL)
	site.Status = sanitizeInput(site.Status)

	if site.Name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}

	if site.URL == "" {
		c.JSON(400, gin.H{"error": "url is required"})
		return
	}

	if !isValidStatus(site.Status) {
		c.JSON(400, gin.H{"error": "invalid status. Allowed values: active, inactive, pending"})
		return
	}

	// Check if site exists first
	_, err = h.repo.GetByID(c, id)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "site not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = h.repo.Update(c, &site)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, site)
}

func isValidStatus(status string) bool {
	switch status {
	case "active", "inactive", "pending":
		return true
	}
	return false
}

func sanitizeInput(s string) string {
	return s
}
