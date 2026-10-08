package website

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo        *Repository
	nginxService *NginxConfigService
}

func NewHandler(repo *Repository, nginxService *NginxConfigService) *Handler {
	return &Handler{
		repo:         repo,
		nginxService: nginxService,
	}
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

	if site.Port == 0 {
		site.Port = 8080
	}

	// Generate default nginx config
	nginxConfig := generateDefaultNginxConfig(site.URL, site.Port)

	err := h.repo.Create(c, &site)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Write nginx config to file
	if err := h.nginxService.CreateConfig(site.ID, nginxConfig); err != nil {
		c.JSON(500, gin.H{"error": "site created but failed to write nginx config: " + err.Error()})
		return
	}

	c.JSON(201, site)
}

func (h *Handler) DeleteSite(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// Delete nginx config file first
	_ = h.nginxService.DeleteConfig(id)

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

func (h *Handler) GetNginxConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	_, err := h.repo.GetByID(c, id)
	if err != nil {
		c.JSON(404, gin.H{"error": "site not found"})
		return
	}

	config, err := h.nginxService.ReadConfig(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"nginx_config": config})
}

func (h *Handler) UpdateNginxConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		NginxConfig string `json:"nginx_config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := h.repo.GetByID(c, id)
	if err != nil {
		c.JSON(404, gin.H{"error": "site not found"})
		return
	}

	if err := h.nginxService.UpdateConfig(id, req.NginxConfig); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "nginx config updated"})
}

func generateDefaultNginxConfig(domain string, port int) string {
	if domain == "" {
		domain = "_"
	}
	return `server {
    listen 80;
    server_name ` + domain + `;

    location / {
        proxy_pass http://127.0.0.1:` + strconv.Itoa(port) + `;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
`
}
