package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"deadmandb/internal/state"
)

type Server struct {
	engine  *gin.Engine
	stateDB string
}

func NewServer(stateDBPath string) *Server {
	r := gin.Default()

	s := &Server{
		engine:  r,
		stateDB: stateDBPath,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	api := s.engine.Group("/api/v1")
	{
		api.GET("/health", s.handleHealth)
		api.GET("/snapshots", s.handleListSnapshots)
		// api.POST("/backups", s.handleTriggerBackup) // Example for future
		// api.POST("/snapshots/:id/restore", s.handleRestore) // Example for future
	}
}

func (s *Server) Start(port int) error {
	return s.engine.Run(fmt.Sprintf(":%d", port))
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

func (s *Server) handleListSnapshots(c *gin.Context) {
	db, err := state.InitDB(s.stateDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to state database"})
		return
	}
	defer db.Close()

	var snapshots []state.Snapshot
	err = db.Select(&snapshots, "SELECT * FROM snapshots ORDER BY timestamp DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch snapshots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": snapshots,
	})
}
