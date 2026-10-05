package main

import (
	"github.com/MetallIdea/adpanel/server/internal/db"
	"github.com/MetallIdea/adpanel/server/internal/router"
)

func main() {
	err := db.Init()
	if err != nil {
		panic(err)
	}

	r := router.SetupRouter()
	r.Run(":8080")
}
