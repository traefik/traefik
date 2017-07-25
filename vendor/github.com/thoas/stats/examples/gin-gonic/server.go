package main

import (
	"github.com/gin-gonic/gin"
	"github.com/thoas/stats"
	"net/http"
)

// Stats provides response time, status code count, etc.
var Stats = stats.New()

func main() {
	r := gin.New()

	r.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			beginning, recorder := Stats.Begin(c.Writer)
			c.Next()
			Stats.End(beginning, recorder)
		}
	}())

	r.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, Stats.Data())
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hello": "world"})
	})

	r.Run(":3001")
}
