package app

import (
	"log"

	"github.com/A7bari/RunWave/internal/db"
	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/gin-gonic/gin"
)

type ServerOpts struct {
	Queues map[string]taskqueue.TaskQueue
}
type Server struct {
	router *gin.Engine
}

func NewServer(opts ServerOpts) *Server {
	r := gin.Default()

	st := db.GetInMemStore()

	// Middleware to inject TaskQueue into context
	r.Use(func(c *gin.Context) {
		// Inject TaskQueues into context
		for lang, q := range opts.Queues {
			c.Set(lang, q)
		}
		c.Set("store", st)
		c.Next()
	})

	server := &Server{
		router: r,
	}

	// Register routes from routes.go
	RegisterRoutes(server.router)

	return server
}

func (s *Server) Start() {
	log.Fatal(s.router.Run(":8080"))
}
