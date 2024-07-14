package route

import (
	"github.com/gin-gonic/gin"
	"github.com/marktran77/go/handlers"
)

func SetUpRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/albums", handlers.GetAlbums)
	router.GET("/albums/:id", handlers.GetAlbumByID)
	router.POST("/albums", handlers.PostAlbums)
	return router
}
