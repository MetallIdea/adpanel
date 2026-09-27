package router

import (
	"github.com/MetallIdea/adpanel/server/internal/auth"
	"github.com/MetallIdea/adpanel/server/internal/command"
	"github.com/MetallIdea/adpanel/server/internal/db"
	"github.com/MetallIdea/adpanel/server/internal/webservice"
	"github.com/MetallIdea/adpanel/server/internal/website"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	siteRepo := website.NewRepository(db.DB)
	siteHandler := website.NewHandler(siteRepo)

	serviceRepo := webservice.NewRepository(db.DB)
	serviceHandler := webservice.NewHandler(serviceRepo)

	authHandler := auth.NewHandler()

	commandHandler := command.NewHandler()

	r := gin.Default()

	authorized := r.Group("/api")
	authorized.Use(auth.AuthMiddleware())

	authorized.GET("/sites", siteHandler.GetSites)
	authorized.GET("/sites/:id", siteHandler.GetSiteByID)
	authorized.GET("/sites/:id/services", serviceHandler.GetServicesBySiteID)
	authorized.POST("/sites", siteHandler.CreateSite)
	authorized.PUT("/sites/:id", siteHandler.UpdateSite)
	authorized.DELETE("/sites/:id", siteHandler.DeleteSite)

	authorized.GET("/services", serviceHandler.GetServices)
	authorized.POST("/services", serviceHandler.CreateService)

	r.POST("/api/login", authHandler.Login)

	authorized.POST("/execute", commandHandler.Execute)

	return r
}
