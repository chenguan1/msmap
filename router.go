package main

import (
	"runtime"

	"github.com/chenguan1/msmap/mswrap"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	if runtime.GOOS == "windows" {
		gin.DisableConsoleColor()
	}
	r := gin.Default()

	r.GET("/", hello)

	// api
	api := r.Group("/api")
	{
		// v1 : /api/v1
		v1 := api.Group("/v1")
		{
			// ping : /api/v1/ping
			v1.GET("/ping", ping)

			// mapserver : /api/v1/ms
			v1.Any("/ms", mswrap.HandleMapServ)

			// dataset : /api/v1/dataset
			dt := v1.Group("/dataset")
			{
				dt.GET("/", listDataset)
				dt.POST("/upload", uploadData)
				dt.POST("/", createDataset)
				dt.GET("/:id/wms", wmsDataset)
				dt.GET("/:id/xyz", xyzDataset)
			}
		}
	}

	return r
}
