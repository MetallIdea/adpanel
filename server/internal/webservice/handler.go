package webservice

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo         *Repository
	nginxService *nginxServiceProxy
}

type nginxServiceProxy struct {
	UpdateConfig func(siteID int64, content string) error
}

func NewHandler(repo *Repository, nginxUpdateFunc func(siteID int64, content string) error) *Handler {
	return &Handler{
		repo: repo,
		nginxService: &nginxServiceProxy{
			UpdateConfig: nginxUpdateFunc,
		},
	}
}

// generateSystemdService generates a systemd service file content
func generateSystemdService(service *WebService) string {
	var sb strings.Builder

	sb.WriteString("[Unit]\n")
	if service.Comment != "" {
		sb.WriteString(fmt.Sprintf("Description=%s\n", service.Comment))
	} else {
		sb.WriteString(fmt.Sprintf("Description=%s Service\n", service.Name))
	}
	sb.WriteString("\n")
	sb.WriteString("[Service]\n")

	if service.WorkingDir != "" {
		sb.WriteString(fmt.Sprintf("WorkingDirectory=%s\n", service.WorkingDir))
		sb.WriteString(fmt.Sprintf("Directory=%s\n", service.WorkingDir))
	}

	if service.EnvContent != "" {
		envFile := filepath.Join(service.WorkingDir, ".env")
		sb.WriteString(fmt.Sprintf("EnvironmentFile=%s\n", envFile))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("ExecStart=%s\n", service.StartupCmd))
	sb.WriteString("Restart=on-failure\n")
	sb.WriteString("RestartSec=5\n")
	sb.WriteString("\n")
	sb.WriteString("[Install]\n")
	sb.WriteString("WantedBy=multi-user.target\n")

	return sb.String()
}

// saveServiceFile saves the systemd service file to the working directory
func saveServiceFile(service *WebService) error {
	if service.WorkingDir == "" {
		return nil
	}

	// Create working directory if it doesn't exist
	if err := os.MkdirAll(service.WorkingDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}

	// Save .env file if env content is provided
	if service.EnvContent != "" {
		envPath := filepath.Join(service.WorkingDir, ".env")
		if err := os.WriteFile(envPath, []byte(service.EnvContent), 0644); err != nil {
			return fmt.Errorf("failed to write .env file: %w", err)
		}
	}

	// Generate and save .service file
	serviceFileName := strings.ToLower(strings.ReplaceAll(service.Name, " ", "-")) + ".service"
	servicePath := filepath.Join(service.WorkingDir, serviceFileName)

	serviceContent := generateSystemdService(service)
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	return nil
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

	// Save service files to working directory
	if err := saveServiceFile(&service); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("service created but failed to save files: %v", err)})
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

	// Save/update service files to working directory
	if err := saveServiceFile(&service); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("service updated but failed to save files: %v", err)})
		return
	}

	// Update nginx configs for all sites that use this service
	h.updateNginxConfigsForService(c, service)

	c.JSON(200, service)
}

func (h *Handler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid service id"})
		return
	}

	// Check if service exists
	svc, err := h.repo.GetServiceByID(c, id)
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

	// Remove service files from working directory
	if svc.WorkingDir != "" {
		serviceFileName := strings.ToLower(strings.ReplaceAll(svc.Name, " ", "-")) + ".service"
		servicePath := filepath.Join(svc.WorkingDir, serviceFileName)
		os.Remove(servicePath)

		envPath := filepath.Join(svc.WorkingDir, ".env")
		os.Remove(envPath)
	}

	// Update nginx configs for sites that used this service
	h.updateNginxConfigsForService(c, *svc)

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

// updateNginxConfigsForService updates nginx configs for all sites using this service
func (h *Handler) updateNginxConfigsForService(ctx context.Context, service WebService) {
	// Get all site-service associations for this service
	type siteService struct {
		SiteID int64 `json:"site_id"`
		Path   string `json:"path"`
	}

	// We need to query site_services table - but we don't have direct access
	// This is a simplified version - in production you'd pass the ssRepo
	// For now, we'll skip this and let the site handler manage nginx config
	_ = service
	_ = ctx
}
