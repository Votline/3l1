package main

import (
	"os"
	"net"
	"errors"
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
		log: log,
		repo: db.NewRepo(log),
		redisRepo: db.NewRR(log),
	}
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

	sessionKey, err := us.redisRepo.NewSession(id, role)
	if err != nil {
		us.log.Error("Failed to create session key", zap.Error(err))
		return nil, err
	}

	if err := us.repo.AddUser(id, name+email, role, pswd); err != nil {
		us.log.Error("Failed to add user into db", zap.Error(err))
		return nil, err
	}

	return &pb.RegRes{Token: token, SessionKey: sessionKey}, nil
}

func (us *userserver) LogUser(ctx context.Context, req *pb.LogReq) (*pb.LogRes, error) {
	name := req.GetName()
	email := req.GetEmail()
	pswd := req.GetPassword()

	data, err := us.repo.LogUser(name+email, pswd);
	if err != nil {
		us.log.Error("Failed extract data", zap.Error(err))
		return nil, err
	}

	if !crypto.CheckPswd(data.Pswd, pswd) {
		us.log.Error("Failed login user")
		return nil, errors.New("Invalid password")
	}
	
	sessionKey, err := us.redisRepo.NewSession(data.ID, data.Role)
	if err != nil {
		us.log.Error("Failed to create session key", zap.Error(err))
		return nil, err
	}

	token, err := crypto.GenJWT(data.ID, data.Role)
	if err != nil {
		us.log.Error("Failed create jwt token", zap.Error(err))
		return nil, err
	}

	return &pb.LogRes{Token: token, SessionKey: sessionKey}, nil
}

func (us *userserver) ExtJWTData(ctx context.Context, req *pb.ExtJWTDataReq) (*pb.ExtJWTDataRes, error ) {
	sk := req.GetSessionKey()
	tokenString := req.GetToken()
	
	data, err := crypto.ExtJWT(tokenString)
	id, role := data.UserID, data.Role
	if id != "" && role != "" && err != nil {
		if err := us.redisRepo.Validate(id, role, sk); err != nil {
			us.log.Error("Failed to falidate data", zap.Error(err))
			return nil, err
		}
		tokenString, err = crypto.GenJWT(id, role)
		if err != nil {
			us.log.Error("Failed to create new JWT token", zap.Error(err))
			return nil, err
		}
	} else if err != nil {
		us.log.Error("Failed to extract any data from JWT token", zap.Error(err))
		return nil, err
	}

	return &pb.ExtJWTDataRes{
		Role: data.Role,
		UserId: data.UserID,
		Token: tokenString,
	}, nil
}

func (us *userserver) DelUser(ctx context.Context, req *pb.DelUserReq) (*pb.DelUserRes, error) {
	role := req.GetRole()
	userID := req.GetUserId()
	delUserID := req.GetDelUserId()
	sk := req.GetSessionKey()

	if err := us.redisRepo.DelSession(sk); err != nil {
		us.log.Error("Failed to delete user's session key", zap.Error(err))
		return nil, err
	}

	if err := us.repo.DelUser(userID, role, delUserID); err != nil {
		us.log.Error("Failed to delete user", zap.Error(err))
		return nil, err
	}

	return &pb.DelUserRes{}, nil
}
