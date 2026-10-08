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

	siteHandler := website.NewHandler(siteRepo, nginxService)

	serviceRepo := webservice.NewRepository(db.DB)
	serviceHandler := webservice.NewHandler(serviceRepo)

	authHandler := auth.NewHandler()

	commandHandler := command.NewHandler()

	serverHandler := server.NewHandler()

	r := gin.Default()

	authorized := r.Group("/api")
	authorized.Use(auth.AuthMiddleware())

	authorized.GET("/sites", siteHandler.GetSites)
	authorized.GET("/sites/:id", siteHandler.GetSiteByID)
	authorized.GET("/sites/:id/services", serviceHandler.GetServicesBySiteID)
	authorized.POST("/sites", siteHandler.CreateSite)
	authorized.PUT("/sites/:id", siteHandler.UpdateSite)
	authorized.DELETE("/sites/:id", siteHandler.DeleteSite)
	authorized.GET("/sites/:id/nginx-config", siteHandler.GetNginxConfig)
	authorized.PUT("/sites/:id/nginx-config", siteHandler.UpdateNginxConfig)

	authorized.GET("/services", serviceHandler.GetServices)
	authorized.POST("/services", serviceHandler.CreateService)

	r.POST("/api/login", authHandler.Login)

	authorized.POST("/execute", commandHandler.Execute)

	authorized.GET("/server/status", serverHandler.GetStatus)

	return r
}
