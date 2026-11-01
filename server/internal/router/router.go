package router

import (
	"fmt"
	"os"
	"strconv"
	"time"

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

	// Load command configuration.
	rules, err := command.LoadWhitelist()
	if err != nil {
		panic(fmt.Sprintf("invalid command whitelist configuration: %v", err))
	}

	defaultTimeout := 30 * time.Second
	if t := os.Getenv("COMMAND_DEFAULT_TIMEOUT"); t != "" {
		d, err := time.ParseDuration(t)
		if err != nil {
			panic(fmt.Sprintf("invalid COMMAND_DEFAULT_TIMEOUT: %v", err))
		}
		if d <= 0 {
			panic("COMMAND_DEFAULT_TIMEOUT must be positive")
		}
		defaultTimeout = d
	}

	maxOutput := 65536 // 64 KiB default.
	if m := os.Getenv("COMMAND_MAX_OUTPUT"); m != "" {
		n, err := command.ParseMaxOutput(m)
		if err != nil {
			panic(fmt.Sprintf("invalid COMMAND_MAX_OUTPUT: %v", err))
		}
		maxOutput = n
	}

	rateLimit := 10.0
	if value := os.Getenv("COMMAND_RATE_LIMIT"); value != "" {
		rateLimit, err = parseRateLimit(value)
		if err != nil {
			panic(fmt.Sprintf("invalid COMMAND_RATE_LIMIT: %v", err))
		}
	}

	if err := command.ValidateConfig(
		defaultTimeout,
		maxOutput,
		rateLimit,
		rules,
	); err != nil {
		panic(fmt.Sprintf("invalid command configuration: %v", err))
	}

	commandHandler := command.NewHandler(
		rules,
		rateLimit,
		defaultTimeout,
		maxOutput,
	)

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
	authorized.GET("/services/:id", serviceHandler.GetServiceByID)
	authorized.POST("/services", serviceHandler.CreateService)
	authorized.PUT("/services/:id", serviceHandler.UpdateService)
	authorized.DELETE("/services/:id", serviceHandler.DeleteService)

	r.POST("/api/login", authHandler.Login)

	authorized.POST("/execute", commandHandler.Execute)

	authorized.GET("/server/status", serverHandler.GetStatus)

	return r
}

func parseRateLimit(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if v <= 0 {
		return 0, fmt.Errorf("rate limit must be positive, got %f", v)
	}
	return v, nil
}
