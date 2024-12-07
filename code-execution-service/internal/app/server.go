package app

import (
	"log"

	"github.com/A7bari/RunWave/internal"
	"github.com/gin-gonic/gin"
)

type Server struct {
	podManager *internal.PodManager
	router     *gin.Engine
}

func NewServer(podManager *internal.PodManager) *Server {
	server := &Server{
		podManager: podManager,
		router:     gin.Default(),
	}

	// Register routes from routes.go
	RegisterRoutes(server.router, podManager)

	return server
}

func (s *Server) Start() {
	log.Fatal(s.router.Run(":8080"))
}
