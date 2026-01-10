package gapi

import (
	"context"
	"fmt"

	db "github.com/chuuch/go-banking/db/sqlc"
	"github.com/chuuch/go-banking/pb"
	"github.com/chuuch/go-banking/token"
	"github.com/chuuch/go-banking/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

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
	return server, nil
}

func (s *Server) CreateUser(
	ctx context.Context,
	req *pb.CreateUserRequest,
) (*pb.CreateUserResponse, error) {
	return &pb.CreateUserResponse{}, nil
}

func (s *Server) LoginUser(
	ctx context.Context,
	req *pb.LoginUserRequest,
) (*pb.LoginUserResponse, error) {
	return &pb.LoginUserResponse{}, nil
}
