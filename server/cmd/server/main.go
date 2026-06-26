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

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	repo := website.NewRepository(db.DB)
	handler := website.NewHandler(repo)

	r.GET("/sites", handler.GetSites)

	r.Run(":8080")
}
