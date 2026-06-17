package server

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed dashboard_root.html
var dashboardRootHTML []byte

// serveDashboardRoot serves the built-in HTML dashboard for browser requests at "/".
func (s *Server) serveDashboardRoot(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", dashboardRootHTML)
}
