package website

import (
	"fmt"
	"os"
	"path/filepath"
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

func (s *NginxConfigService) initDir() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create nginx config dir: %w", err)
	}
	return nil
}

func (s *NginxConfigService) GetConfigPath(siteID int64) string {
	return filepath.Join(s.configDir, fmt.Sprintf("site_%d.conf", siteID))
}

func (s *NginxConfigService) CreateConfig(siteID int64, content string) error {
	if err := s.initDir(); err != nil {
		return err
	}

	path := s.GetConfigPath(siteID)
	return os.WriteFile(path, []byte(content), 0644)
}

func (s *NginxConfigService) UpdateConfig(siteID int64, content string) error {
	path := s.GetConfigPath(siteID)
	return os.WriteFile(path, []byte(content), 0644)
}

func (s *NginxConfigService) ReadConfig(siteID int64) (string, error) {
	path := s.GetConfigPath(siteID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read nginx config: %w", err)
	}
	return string(data), nil
}

func (s *NginxConfigService) DeleteConfig(siteID int64) error {
	path := s.GetConfigPath(siteID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete nginx config: %w", err)
	}
	return nil
}
