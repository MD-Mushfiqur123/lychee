package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/MD-Mushfiqur123/lychee/api"
)

// CreateRouteHandler registers or updates a virtual model route.
func (s *Server) CreateRouteHandler(c *gin.Context) {
	if s.modelRouter == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model router not initialized"})
		return
	}

	var req api.RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "route name is required"})
		return
	}

	if len(req.Endpoints) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoints list cannot be empty"})
		return
	}

	route := api.ModelRoute{
		Name:      req.Name,
		Endpoints: req.Endpoints,
		Strategy:  req.Strategy,
	}

	if err := s.modelRouter.AddRoute(route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, route)
}

// ListRoutesHandler lists all virtual model routes.
func (s *Server) ListRoutesHandler(c *gin.Context) {
	if s.modelRouter == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model router not initialized"})
		return
	}

	routes := s.modelRouter.ListRoutes()
	if routes == nil {
		routes = []api.ModelRoute{}
	}

	c.JSON(http.StatusOK, routes)
}

// DeleteRouteHandler removes a virtual model route.
func (s *Server) DeleteRouteHandler(c *gin.Context) {
	if s.modelRouter == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model router not initialized"})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if err := s.modelRouter.RemoveRoute(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// RouteStatusHandler queries the status of a virtual model route.
func (s *Server) RouteStatusHandler(c *gin.Context) {
	if s.modelRouter == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model router not initialized"})
		return
	}
	name := c.Param("name")
	status, err := s.modelRouter.GetRouteStatus(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}
