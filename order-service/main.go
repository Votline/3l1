package main

import (
	"os"
	"net"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"github.com/google/uuid"

	"orders/internal/db"
	pb "github.com/Votline/3l1/protos/generated-order"
)

type orderservice struct {
	log *zap.Logger
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
	s.Serve(lis)
}

func (os *orderservice) AddOrder(ctx context.Context, req *pb.AddOrderReq) (*pb.AddOrderRes, error) {
	order := &db.Order{
		ID: uuid.New().String(),
		UserID: req.GetUserId(),
		TargetURL: req.GetTargetUrl(),
		ServiceURL: req.GetServiceUrl(),
		OrderType: req.GetOrderType(),
		Quantity: req.GetQuantity(),
	}
	
	if err := os.repo.AddOrder(order); err != nil {
		os.log.Error("Failed to add order into db", zap.Error(err))
		return nil, err
	}

	return &pb.AddOrderRes{Id: order.ID}, nil
}
