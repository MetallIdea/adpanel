package main

import (
	"github.com/MetallIdea/adpanel/server/internal/db"
	"github.com/MetallIdea/adpanel/server/internal/migration"
)

func main() {
	err := db.Init()
	if err != nil {
		panic(err)
	}
	defer db.DB.Close()

	migrator, err := migration.NewMigrator(db.DB, "migration")
	if err != nil {
		panic(err)
	}

	if err = migrator.Init(); err != nil {
		panic(err)
	}

	if err = migrator.Run(); err != nil {
		panic(err)
	}

	r := setupRouter()
	r.Run(":8080")
}
