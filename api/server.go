package api

import (
	db "github.com/chuuch/go-banking/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  *db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Create account
	router.POST("/accounts", server.createAccount)
	// Get account
	router.GET("/accounts/:id", server.getAccount)
	// Update account (partial)
	router.PATCH("/accounts/:id", server.updateAccount)
	// Delete account
	router.DELETE("/accounts/:id", server.deleteAccount)
	// List accounts (paginated)
	router.GET("/accounts", server.listAccount)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
