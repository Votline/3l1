package main

import (
	"os"
	"net"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"users/internal/db"
	"users/internal/crypto"

	"github.com/google/uuid"
	pb "github.com/Votline/3l1/protos/generated-user"
)

type userserver struct {
	log *zap.Logger
	repo *db.Repo
	pb.UnimplementedUserServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()
	lis, err := net.Listen("tcp", ":"+os.Getenv("US_PORT"))
	if err != nil {
		log.Fatal("Couldn't listen tcp user-service port", zap.Error(err))
	}

	s := grpc.NewServer()
	srv := userserver{log: log, repo: db.NewRepo(log)}
	pb.RegisterUserServiceServer(s, &srv)
	s.Serve(lis)
}

func (us *userserver) HashPswd(ctx context.Context, req *pb.HashReq) (*pb.HashRes, error) {
	pswd := req.GetPassword()
	
	hashed, err := crypto.Hash(pswd)
	if err != nil {
		us.log.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	return &pb.HashRes{PasswordHash: hashed}, nil
}

func (us *userserver) RegUser(ctx context.Context, req *pb.RegReq) (*pb.RegRes, error) {
	name  := req.GetName()
	email := req.GetEmail()
	role  := req.GetRole()
	pswd  := req.GetPasswordHash()
	id    := uuid.New().String()

	token, err := crypto.GenJWT(id, role)
	if err != nil {
		us.log.Error("Failed create jwt token", zap.Error(err))
		return nil, err
	}

	if err := us.repo.AddUser(id, name+email, role, pswd); err != nil {
		us.log.Error("Failed to add user into db", zap.Error(err))
		return nil, err
	}

	return &pb.RegRes{Token: token}, nil
}

func (us *userserver) LogUser(ctx context.Context, req *pb.LogReq) (*pb.LogRes, error) {
	name := req.GetName()
	email := req.GetEmail()
	pswd := req.GetPasswordHash()

	data, err := us.repo.LogUser(name+email, pswd);
	if err != nil {
		us.log.Error("Failed login user", zap.Error(err))
		return nil, err
	}

	token, err := crypto.GenJWT(data.ID, data.Role)
	if err != nil {
		us.log.Error("Failed create jwt token", zap.Error(err))
		return nil, err
	}

	return &pb.LogRes{Token: token}, nil
}
