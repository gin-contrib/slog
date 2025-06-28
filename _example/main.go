package main

import (
	"net/http"

	"github.com/gin-contrib/slog"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	// Add slog middleware with default settings
	r.Use(slog.SetLogger())

	// Example route
	r.GET("/", func(c *gin.Context) {
		slog.Get(c).Info("Hello World!")
		c.String(http.StatusOK, "ok")
	})

	r.Run()
}
