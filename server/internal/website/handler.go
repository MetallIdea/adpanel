package website

import "github.com/gin-gonic/gin"

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) GetSites(c *gin.Context) {
	sites, err := h.repo.GetAll(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, sites)
}
