package main

import (
	"github.com/gin-gonic/gin"
)

// APIServerStart api
func APIServerStart() {
	router := gin.Default()

	router.GET("/ping", ping)

	router.GET("/info", getInfo)

	router.POST("/info", postInfo)

	router.GET("/data", getData)

	router.Run(":90")
}

func ping(context *gin.Context) {
	context.JSON(200, gin.H{
		"message": "pong",
	})
}

func getInfo(context *gin.Context) {
	config := Config{}
	Orm.Get(&config)
	context.JSON(200, config)
}

func postInfo(context *gin.Context) {
	config := Config{}
	context.Bind(&config)
	Orm.Where("id = 1").Update(&config)
	Orm.Get(&config)
	context.JSON(200, config)
}

func getData(context *gin.Context) {
	data := Data{}
	Orm.Desc("timestamp").Get(&data)
	context.JSON(200, data)
}
