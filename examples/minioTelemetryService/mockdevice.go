package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/get_file", func(context *gin.Context) {
		//context.String(200, "test content")
		context.File("test.mp4")
	})

	router.Run("0.0.0.0:12345")
}
