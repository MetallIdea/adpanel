package website

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

func (h *Handler) GetSites(c *gin.Context) {
	sites, err := h.repo.GetAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, sites)
}

func (h *Handler) GetSiteByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	site, err := h.repo.GetByID(c, id)
	if err != nil {
		c.JSON(404, gin.H{"error": "site not found"})
		return
	}

	c.JSON(200, site)
}

func (h *Handler) CreateSite(c *gin.Context) {
	var site Website

	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
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
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := h.repo.Delete(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}

func (h *Handler) UpdateSite(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var site Website
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	site.ID = id

	err := h.repo.Update(c, &site)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, site)
}
