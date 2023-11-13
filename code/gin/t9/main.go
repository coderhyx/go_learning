package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/my-handler", MyHandler)
	r.Run(":8080")
}
