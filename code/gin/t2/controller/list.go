package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListAction(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"create": "success"})
}
