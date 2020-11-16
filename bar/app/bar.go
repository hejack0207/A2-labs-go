package main

import "github.com/gin-gonic/gin"
import "net/http"

func main() {
	r := gin.Default()
	r.GET("/welcome", func(c *gin.Context) {
		name := c.DefaultQuery("name", "Guest")

		c.String(http.StatusOK, "Welcome %s", name)
	})
	r.Run()
}
