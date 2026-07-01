package main

import (
	"github.com/MetallIdea/adpanel/server/internal/db"
	"github.com/MetallIdea/adpanel/server/internal/webservice"
	"github.com/MetallIdea/adpanel/server/internal/website"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	siteRepo := website.NewRepository(db.DB)
	siteHandler := website.NewHandler(siteRepo)

	serviceRepo := webservice.NewRepository(db.DB)
	serviceHandler := webservice.NewHandler(serviceRepo)

	r := gin.Default()

	r.GET("/sites", siteHandler.GetSites)
	r.GET("/sites/:id", siteHandler.GetSiteByID)
	r.GET("/sites/:id/services", serviceHandler.GetServicesBySiteID)
	r.POST("/sites", siteHandler.CreateSite)
	r.PUT("/sites/:id", siteHandler.UpdateSite)
	r.DELETE("/sites/:id", siteHandler.DeleteSite)

	r.GET("/services", serviceHandler.GetServices)
	r.POST("/services", serviceHandler.CreateService)

	return r
}
