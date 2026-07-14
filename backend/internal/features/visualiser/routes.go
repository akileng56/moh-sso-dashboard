package visualiser

import "github.com/gin-gonic/gin"

func RegisterRoutes(visualiser *gin.RouterGroup, handler *Handler) {
	visualiser.GET("/datasets", handler.GetDatasets)
	visualiser.GET("/dataelements", handler.GetDataElements)
	visualiser.POST("/datavalues", handler.GetDataValues)
}
