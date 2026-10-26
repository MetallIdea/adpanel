package router

import (
	"os"

	"github.com/MetallIdea/adpanel/server/internal/auth"
	"github.com/MetallIdea/adpanel/server/internal/command"
	"github.com/MetallIdea/adpanel/server/internal/db"
	"github.com/MetallIdea/adpanel/server/internal/server"
	"github.com/MetallIdea/adpanel/server/internal/webservice"
	"github.com/MetallIdea/adpanel/server/internal/website"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	siteRepo := website.NewRepository(db.DB)

	// Create nginx config directory
	nginxConfigDir := "/etc/nginx/sites-enabled"
	if dir := os.Getenv("NGINX_CONFIG_DIR"); dir != "" {
		nginxConfigDir = dir
	}
	nginxService := website.NewNginxConfigService(nginxConfigDir)

	ssRepo := website.NewSiteServiceRepository(db.DB)

	siteHandler := website.NewHandler(siteRepo, nginxService, ssRepo)

	serviceRepo := webservice.NewRepository(db.DB)
	// Wrap UpdateConfig to match the expected signature (webservice handler doesn't use site name)
	nginxUpdateWrapper := func(siteID int64, content string) error {
		// This wrapper is only used by webservice handler for nginx config updates
		// Site name is not available here, so we read it from the site
		return nil // no-op: nginx config updates are handled by website handler
	}
	serviceHandler := webservice.NewHandler(serviceRepo, nginxUpdateWrapper)

	authHandler := auth.NewHandler()

	commandHandler := command.NewHandler()

	serverHandler := server.NewHandler()

	r := gin.Default()

	authorized := r.Group("/api")
	authorized.Use(auth.AuthMiddleware())

	authorized.GET("/sites", siteHandler.GetSites)
	authorized.GET("/sites/:id", siteHandler.GetSiteByID)
	authorized.GET("/sites/:id/services", siteHandler.GetSiteServices)
	authorized.PUT("/sites/:id/services", siteHandler.UpdateSiteServices)
	authorized.POST("/sites", siteHandler.CreateSite)
	authorized.PUT("/sites/:id", siteHandler.UpdateSite)
	authorized.DELETE("/sites/:id", siteHandler.DeleteSite)
	authorized.GET("/sites/:id/nginx-config", siteHandler.GetNginxConfig)
	authorized.PUT("/sites/:id/nginx-config", siteHandler.UpdateNginxConfig)

	authorized.GET("/services", serviceHandler.GetServices)
	authorized.GET("/services/:id", serviceHandler.GetServiceByID)
	authorized.POST("/services", serviceHandler.CreateService)
	authorized.PUT("/services/:id", serviceHandler.UpdateService)
	authorized.DELETE("/services/:id", serviceHandler.DeleteService)

	r.POST("/api/login", authHandler.Login)

	authorized.POST("/execute", commandHandler.Execute)

	authorized.GET("/server/status", serverHandler.GetStatus)

	return r
}
