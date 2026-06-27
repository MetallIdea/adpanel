package main

import (
	"github.com/MetallIdea/adpanel/server/internal/db"
	"github.com/MetallIdea/adpanel/server/internal/website"
	"github.com/gin-gonic/gin"
)

func main() {
	err := db.Init()
	if err != nil {
		panic(err)
	}

	repo := website.NewRepository(db.DB)
	handler := website.NewHandler(repo)

	r := gin.Default()

	r.GET("/sites", handler.GetSites)
	r.GET("/sites/:id", handler.GetSiteByID)
	r.POST("/sites", handler.CreateSite)
	r.PUT("/sites/:id", handler.UpdateSite)
	r.DELETE("/sites/:id", handler.DeleteSite)

	r.Run(":8080")
}
