package middlewares

import (
	"os"
	"strings"

	"github.com/ThanvirXo/jira-auto-bug-solver/common"
	"github.com/gin-gonic/gin"
)


func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token:=c.GetHeader("Authorization")
		if token==""{
			common.NewResponse(common.UNAUTHORIZED,"Unauthorized").Respond(c)
			return
		}

		if !strings.HasPrefix(token,"Bearer "){
			common.NewResponse(common.UNAUTHORIZED,"Unauthorized").Respond(c)
			return
		}

		token=strings.TrimPrefix(token,"Bearer ")
		token=strings.TrimSpace(token)

		if token==""{
			common.NewResponse(common.UNAUTHORIZED,"Unauthorized").Respond(c)
			return
		}

		if token==os.Getenv("API_KEY"){
			c.Next()
			return
		}

		common.NewResponse(common.UNAUTHORIZED,"Unauthorized").Respond(c)
		
	}
}