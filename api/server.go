package api

import (
	db "github.com/chuuch/go-banking/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// We call gin's binding to access the validator engine and register
	// our custom validCurrency helper variable
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

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

	// Create transfer
	router.POST("/transfers", server.createTransfer)

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
