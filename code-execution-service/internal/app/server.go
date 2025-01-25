package app

import (
	"log"

	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"

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

	st := db.GetPostgresStore()

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

	// Register pprof routes
	RegisterPprofRoutes(server.router)

	return server
}

func (s *Server) Start() {
	log.Fatal(s.router.Run(":8080"))
}

func RegisterPprofRoutes(router *gin.Engine) {
	router.GET("/debug/pprof/", gin.WrapH(http.HandlerFunc(pprof.Index)))
	router.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))
	router.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	router.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))
	router.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	router.GET("/debug/pprof/profile", gin.WrapH(http.HandlerFunc(pprof.Profile)))
	router.POST("/debug/pprof/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
	router.GET("/debug/pprof/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
	router.GET("/debug/pprof/trace", gin.WrapH(http.HandlerFunc(pprof.Trace)))
	router.GET("/debug/pprof/cmdline", gin.WrapH(http.HandlerFunc(pprof.Cmdline)))
}
