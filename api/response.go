package api

import "github.com/gin-gonic/gin"

func SendSuccess(c *gin.Context, data any) {
	c.JSON(200, gin.H{"data": data})
}

func SendError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
