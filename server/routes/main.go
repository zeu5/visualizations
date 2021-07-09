package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/zeu5/visualizations/server/routes/crime"
)

func Initialize(r *gin.Engine) {
	crime.Initialize(r.Group("/crime"))
}
