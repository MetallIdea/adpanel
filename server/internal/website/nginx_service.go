package website

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type NginxConfigService struct {
	configDir string
	mu        sync.RWMutex
}

func NewNginxConfigService(configDir string) *NginxConfigService {
	return &NginxConfigService{
		configDir: configDir,
	}
}

// sanitizeName converts site name to a safe filename
func sanitizeName(name string) string {
	// Replace spaces and special characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	sanitized := re.ReplaceAllString(name, "_")
	// Collapse multiple underscores
	re2 := regexp.MustCompile(`_+`)
	sanitized = re2.ReplaceAllString(sanitized, "_")
	// Trim underscores from edges
	sanitized = strings.Trim(sanitized, "_")
	if sanitized == "" {
		sanitized = "site"
	}
	return sanitized
}

func (s *NginxConfigService) initDir() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create nginx config dir: %w", err)
	}
	return nil
}

// GetConfigPathByID returns the config path using site ID (legacy, for backward compatibility)
func (s *NginxConfigService) GetConfigPath(siteID int64) string {
	return filepath.Join(s.configDir, fmt.Sprintf("site_%d.conf", siteID))
}

// GetConfigPathByName returns the config path using site name
func (s *NginxConfigService) GetConfigPathByName(siteName string) string {
	safeName := sanitizeName(siteName)
	return filepath.Join(s.configDir, fmt.Sprintf("%s.conf", safeName))
}

func (s *NginxConfigService) CreateConfig(siteID int64, siteName string, content string) error {
	if err := s.initDir(); err != nil {
		return err
	}

	path := s.GetConfigPathByName(siteName)
	return os.WriteFile(path, []byte(content), 0644)
}

func (s *NginxConfigService) UpdateConfig(siteID int64, siteName string, content string) error {
	path := s.GetConfigPathByName(siteName)
	return os.WriteFile(path, []byte(content), 0644)
}

func (s *NginxConfigService) ReadConfig(siteID int64) (string, error) {
	// Legacy: try to read by ID first for backward compatibility
	path := s.GetConfigPath(siteID)
	data, err := os.ReadFile(path)
	if err == nil {
		return string(data), nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to read nginx config: %w", err)
	}
	return "", nil
}

func (s *NginxConfigService) DeleteConfig(siteID int64, siteName string) error {
	// Delete by name
	pathByName := s.GetConfigPathByName(siteName)
	if err := os.Remove(pathByName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete nginx config (by name): %w", err)
	}
	// Also try to delete by ID for backward compatibility
	pathByID := s.GetConfigPath(siteID)
	if err := os.Remove(pathByID); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete nginx config (by ID): %w", err)
	}
	return nil
}
