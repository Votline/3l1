package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"users/internal/crypto"
	"users/internal/db"
	gc "users/internal/graceful"

	pb "github.com/Votline/3l1/protos/generated-user"
	"github.com/google/uuid"
)

type userserver struct {
	log       *zap.Logger
	repo      *db.Repo
	redisRepo *db.RedisRepo
	pb.UnimplementedUserServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()
	lis, err := net.Listen("tcp", ":"+os.Getenv("US_PORT"))
	if err != nil {
		log.Fatal("Couldn't listen tcp user-service port", zap.Error(err))
	}

	s := grpc.NewServer()
	srv := userserver{
		log:       log,
		repo:      db.NewRepo(log),
		redisRepo: db.NewRR(log),
	}
	pb.RegisterUserServiceServer(s, &srv)

	go s.Serve(lis)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit
	log.Warn("Shutdown signal received")
	gracefulShutdown(s, srv, log)
}

func gracefulShutdown(s *grpc.Server, srv userserver, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("Shutting down gRPC server")
	if err := gc.Shutdown(
		func() error { s.Stop(); return nil }, ctx); err != nil {
		log.Error("gRPC server shutdown error", zap.Error(err))
	}

	log.Info("Shutting down postgreSQL")
	if err := srv.repo.Stop(ctx); err != nil {
		log.Error("Postgres shutdown error", zap.Error(err))
	}

	log.Info("Shutting down redis")
	if err := srv.redisRepo.Stop(ctx); err != nil {
		log.Error("Redis shutdown error", zap.Error(err))
	}
}

func (us *userserver) RegUser(ctx context.Context, req *pb.RegReq) (*pb.RegRes, error) {
	const op = "UserService.ReqUser"

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%s validate: %w", op, err)
	}

	name := req.GetName()
	email := req.GetEmail()
	role := req.GetRole()
	pswd := req.GetPassword()
	id := uuid.New().String()

	hashed, err := crypto.Hash(pswd)
	if err != nil {
		return nil, fmt.Errorf("%s: hash password: %w", op, err)
	}

	token, err := crypto.GenJWT(id, role)
	if err != nil {
		return nil, fmt.Errorf("%s: generate jwt: %w", op, err)
	}

	sessionKey, err := us.redisRepo.NewSession(id, role)
	if err != nil {
		return nil, fmt.Errorf("%s: new session: %w", op, err)
	}

	if err := us.repo.AddUser(id, name+email, role, hashed); err != nil {
		return nil, fmt.Errorf("%s: add user: %w", op, err)
	}

	return &pb.RegRes{Token: token, SessionKey: sessionKey}, nil
}

func (us *userserver) LogUser(ctx context.Context, req *pb.LogReq) (*pb.LogRes, error) {
	const op = "UserService.LogUser"

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%s validate: %w", op, err)
	}

	name := req.GetName()
	email := req.GetEmail()
	pswd := req.GetPassword()

	data, err := us.repo.LogUser(name+email, pswd)
	if err != nil {
		return nil, fmt.Errorf("%s: login user: %w", op, err)
	}

	if !crypto.CheckPswd(data.Pswd, pswd) {
		return nil, fmt.Errorf("%s: check password: %s", op, "Invalid password")
	}

	sessionKey, err := us.redisRepo.NewSession(data.ID, data.Role)
	if err != nil {
		return nil, fmt.Errorf("%s: new session: %w", op, err)
	}

	token, err := crypto.GenJWT(data.ID, data.Role)
	if err != nil {
		return nil, fmt.Errorf("%s: generate jwt: %w", op, err)
	}

	return &pb.LogRes{Token: token, SessionKey: sessionKey}, nil
}

func (us *userserver) ExtJWTData(ctx context.Context, req *pb.ExtJWTDataReq) (*pb.ExtJWTDataRes, error) {
	const op = "UserService.ExtJWTData"

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%s validate: %w", op, err)
	}

	sk := req.GetSessionKey()
	tokenString := req.GetToken()

	data, err := crypto.ExtJWT(tokenString)
	id, role := data.UserID, data.Role
	if id != "" && role != "" && err == nil {
		if err := us.redisRepo.Validate(id, role, sk); err != nil {
			return nil, fmt.Errorf("%s: validate: %w", op, err)
		}
		tokenString, err = crypto.GenJWT(id, role)
		if err != nil {
			return nil, fmt.Errorf("%s: generate jwt: %w", op, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("%s: extract jwt: %w", op, err)
	}

	return &pb.ExtJWTDataRes{
		Role:   data.Role,
		UserId: data.UserID,
		Token:  tokenString,
	}, nil
}

func (us *userserver) DelUser(ctx context.Context, req *pb.DelUserReq) (*pb.DelUserRes, error) {
	const op = "UserService.DelUser"

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%s validate: %w", op, err)
	}

	role := req.GetRole()
	userID := req.GetUserId()
	delUserID := req.GetDelUserId()
	sk := req.GetSessionKey()

	if err := us.redisRepo.DelSession(sk); err != nil {
		return nil, fmt.Errorf("%s: delete session: %w", op, err)
	}

	if err := us.repo.DelUser(userID, role, delUserID); err != nil {
		return nil, fmt.Errorf("%s: delete user: %w", op, err)
	}

	return &pb.DelUserRes{}, nil
}
