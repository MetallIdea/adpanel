package website

import "github.com/gin-gonic/gin"

func GetSites(c *gin.Context) {
	sites := []Website{
		{
			ID:   1,
			Name: "Google",
			URL:  "https://google.com",
		},
		{
			ID:   2,
			Name: "GitHub",
			URL:  "https://github.com",
		},
	}

	c.JSON(200, sites)
}
