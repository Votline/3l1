package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/Votline/3l1/protos/generated-order"
	"orders/internal/db"
	gc "orders/internal/graceful"
)

type orderservice struct {
	log  *zap.Logger
	repo *db.Repo
	pb.UnimplementedOrderServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()
	lis, err := net.Listen("tcp", ":"+os.Getenv("OS_PORT"))
	if err != nil {
		log.Fatal("Couldn't listen tcp order-service port", zap.Error(err))
	}

	s := grpc.NewServer()
	srv := orderservice{log: log, repo: db.NewRepo(log)}
	pb.RegisterOrderServiceServer(s, &srv)
	go s.Serve(lis)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Warn("Shutdown signal received")
	gracefulShutdown(s, srv, log)
}
func gracefulShutdown(s *grpc.Server, srv orderservice, log *zap.Logger) {
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
}

func (os *orderservice) AddOrder(ctx context.Context, req *pb.AddOrderReq) (*pb.AddOrderRes, error) {
	order := &db.Order{
		ID:         uuid.New().String(),
		UserID:     req.GetUserId(),
		UserRl:     req.GetUserRole(),
		TargetURL:  req.GetTargetUrl(),
		ServiceURL: req.GetServiceUrl(),
		OrderType:  req.GetOrderType(),
		Quantity:   req.GetQuantity(),
	}

	if err := os.repo.AddOrder(order); err != nil {
		os.log.Error("Failed to add order into db", zap.Error(err))
		return nil, err
	}

	return &pb.AddOrderRes{Id: order.ID}, nil
}

func (os *orderservice) OrderInfo(ctx context.Context, req *pb.OrderInfoReq) (*pb.OrderInfoRes, error) {
	id := req.GetId()
	userID := req.GetUserId()

	order, err := os.repo.OrderInfo(id, userID)
	if err != nil {
		os.log.Error("Failed to extract data", zap.Error(err))
		return nil, err
	}

	return &pb.OrderInfoRes{
		UserId:     order.UserID,
		UserRole:   order.UserRl,
		Status:     order.Status,
		TargetUrl:  order.TargetURL,
		ServiceUrl: order.ServiceURL,
		OrderType:  order.OrderType,
		CreatedAt:  timestamppb.New(order.CreatedAt),
		UpdatedAt:  timestamppb.New(order.UpdatedAt),
	}, nil
}

func (os *orderservice) DelOrder(ctx context.Context, req *pb.DelOrderReq) (*pb.DelOrderRes, error) {
	id := req.GetId()
	userID := req.GetUserId()
	role := req.GetRole()

	if err := os.repo.DelOrder(id, userID, role); err != nil {
		os.log.Error("Failed to delete order", zap.Error(err))
		return nil, err
	}

	return &pb.DelOrderRes{}, nil
}
