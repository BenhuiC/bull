package {{ .ProjectName }}

import (
	"github.com/gin-gonic/gin"
)

type Service struct {
}

func (s *Service) Ping(c *gin.Context) {
	c.JSON(200, "pong")
}


