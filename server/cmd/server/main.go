package main

import (
	"github.com/MetallIdea/adpanel/server/internal/db"
)

func main() {
	err := db.Init()
	if err != nil {
		panic(err)
	}

	r := setupRouter()
	r.Run(":8080")
}
