package api

import (
	"fmt"

	db "github.com/chuuch/go-banking/db/sqlc"
	"github.com/chuuch/go-banking/token"
	"github.com/chuuch/go-banking/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	// We call gin's binding to access the validator engine and register
	// our custom validCurrency helper variable
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {

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

	// Create transfer
	router.POST("/transfers", server.createTransfer)

	// Create user
	router.POST("/users", server.createUser)
	// Log user in
	router.POST("/users/login", server.loginUser)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
