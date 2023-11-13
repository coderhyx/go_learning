package main

import (
	"go_learning/gin/t2/router"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	router.GoodsLoad(r)
	router.PayLoad(r)
	r.Run(":8000")
}
