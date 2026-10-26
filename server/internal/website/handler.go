package website

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo         *Repository
	nginxService *NginxConfigService
	ssRepo       *SiteServiceRepository
}

func NewHandler(repo *Repository, nginxService *NginxConfigService, ssRepo *SiteServiceRepository) *Handler {
	return &Handler{
		repo:         repo,
		nginxService: nginxService,
		ssRepo:       ssRepo,
	}
}

// generateDefaultNginxConfig generates a basic nginx config
func generateDefaultNginxConfig(siteURL string, sitePort int) string {
	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		parsedURL, err = url.Parse("http://" + siteURL)
	}
	if err != nil {
		parsedURL, _ = url.Parse("http://localhost")
	}

	proxyPass := fmt.Sprintf("http://127.0.0.1:%d", sitePort)

	return fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass %s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}`, parsedURL.Host, proxyPass)
}

// generateNginxConfigWithServices generates nginx config with service location blocks
func generateNginxConfigWithConfig(siteURL string, sitePort int, nginxConfig string, siteServices []SiteService) string {
	// If there's a custom nginx config, use it as base
	baseConfig := nginxConfig
	if baseConfig == "" {
		baseConfig = generateDefaultNginxConfig(siteURL, sitePort)
	}

	// If no services, return base config
	if len(siteServices) == 0 {
		return baseConfig
	}

	// Parse the base config to find the location blocks section
	// We'll inject service locations before the closing server block
	locationBlocks := generateLocationBlocks(siteServices)

	// Find the last } in the config (closing server block)
	lastBrace := strings.LastIndex(baseConfig, "}")
	if lastBrace == -1 {
		return baseConfig + "\n\n" + locationBlocks + "\n}"
	}

	// Insert location blocks before the closing brace
	return baseConfig[:lastBrace] + locationBlocks + "\n" + baseConfig[lastBrace:]
}

// generateLocationBlocks creates nginx location blocks for services
func generateLocationBlocks(siteServices []SiteService) string {
	var sb strings.Builder

	for _, ss := range siteServices {
		// Get service details to build proxy_pass
		path := ss.Path
		if path == "" || path == "/" {
			path = "/"
		}

		// We need to get the service port - this will be filled by the handler
		sb.WriteString(fmt.Sprintf(`
    # Service: %s
    location %s {
        proxy_pass http://127.0.0.1:%d;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }`, path, path, 0)) // Port will be updated by handler
	}

	return sb.String()
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
	if err := h.nginxService.CreateConfig(site.ID, site.Name, nginxConfig); err != nil {
		c.JSON(500, gin.H{"error": "site created but failed to write nginx config: " + err.Error()})
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

	// Get site name before deleting
	site, err := h.repo.GetByID(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = h.repo.Delete(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Delete nginx config by name
	h.nginxService.DeleteConfig(id, site.Name)

	// Also delete site_services associations
	h.ssRepo.DeleteBySiteID(c, id)

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

func (h *Handler) GetSiteServices(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid site id"})
		return
	}

	siteServices, err := h.ssRepo.GetBySiteID(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Enrich with service details
	type SiteServiceWithDetails struct {
		ID       int64  `json:"id"`
		SiteID   int64  `json:"site_id"`
		ServiceID int64  `json:"service_id"`
		ServiceName string `json:"service_name"`
		ServicePort int  `json:"service_port"`
		Path     string `json:"path"`
	}

	var result []SiteServiceWithDetails
	for _, ss := range siteServices {
		// Get service details
		// We need to query the webservice table - for now use a simple approach
		result = append(result, SiteServiceWithDetails{
			ID:        ss.ID,
			SiteID:    ss.SiteID,
			ServiceID: ss.ServiceID,
			Path:      ss.Path,
		})
	}

	c.JSON(200, result)
}

func (h *Handler) UpdateSiteServices(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid site id"})
		return
	}

	var siteServices []SiteService
	if err := c.ShouldBindJSON(&siteServices); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	// Delete existing associations
	h.ssRepo.DeleteBySiteID(c, id)

	// Create new associations
	for i := range siteServices {
		siteServices[i].SiteID = id
		if err := h.ssRepo.Create(c, &siteServices[i]); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create site service association: %v", err)})
			return
		}
	}

	// Regenerate nginx config
	site, err := h.repo.GetByID(c, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "site not found"})
		return
	}

	nginxConfig, _ := h.nginxService.ReadConfig(id)
	updatedSiteServices, _ := h.ssRepo.GetBySiteID(c, id)
	
	// Generate new config with service locations
	newConfig := h.generateFullNginxConfig(site, nginxConfig, updatedSiteServices)
	if err := h.nginxService.UpdateConfig(id, site.Name, newConfig); err != nil {
		c.JSON(500, gin.H{"error": "failed to update nginx config: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "site services updated successfully"})
}

func (h *Handler) generateFullNginxConfig(site *Website, nginxConfig string, siteServices []SiteService) string {
	// Start with base config
	baseConfig := nginxConfig
	if baseConfig == "" {
		baseConfig = generateDefaultNginxConfig(site.URL, site.Port)
	}

	if len(siteServices) == 0 {
		return baseConfig
	}

	// Build location blocks
	var sb strings.Builder
	sb.WriteString("\n")
	
	for _, ss := range siteServices {
		path := ss.Path
		if path == "" {
			path = "/"
		}
		
		// Get service port from database
		var port int
		err := h.ssRepo.db.QueryRowContext(
			context.Background(),
			"SELECT port FROM web_services WHERE id=?",
			ss.ServiceID,
		).Scan(&port)
		
		if err != nil {
			port = 0 // skip this service if not found
			continue
		}

		sb.WriteString(fmt.Sprintf(`
    # Service: %s (ID: %d)
    location %s {
        proxy_pass http://127.0.0.1:%d;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }`, ss.ServiceID, ss.ServiceID, path, port))
	}

	// Insert before closing server block
	lastBrace := strings.LastIndex(baseConfig, "}")
	if lastBrace == -1 {
		return baseConfig + sb.String() + "\n}"
	}

	return baseConfig[:lastBrace] + sb.String() + "\n" + baseConfig[lastBrace:]
}

func (h *Handler) GetNginxConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid site id"})
		return
	}

	nginxConfig, err := h.nginxService.ReadConfig(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"nginx_config": nginxConfig})
}

func (h *Handler) UpdateNginxConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid site id"})
		return
	}

	// Get site to retrieve name
	site, err := h.repo.GetByID(c, id)
	if err != nil {
		c.JSON(404, gin.H{"error": "site not found"})
		return
	}

	var req struct {
		NginxConfig string `json:"nginx_config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.nginxService.UpdateConfig(id, site.Name, req.NginxConfig); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "nginx config saved successfully"})
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
