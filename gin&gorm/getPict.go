package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化引擎
	router := gin.Default()
	// 注册一个路由和处理函数
	//router.Static("/assets", "./assets")
	router.StaticFS("/static", http.Dir("assets"))
	router.Run(":9205")
}
