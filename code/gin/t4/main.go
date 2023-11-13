package main

import (
	"go_learning/gin/t4/apps/inventory"
	"go_learning/gin/t4/apps/order"
	"go_learning/gin/t4/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	routers.Include(order.Routers, inventory.Routers)
	routers.Init(r)
	r.Run(":8000")
}
