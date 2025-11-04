package main

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/Votline/3l1/protos/generated-user"
)

type userserver struct {
	pb.UnimplementedProjectServiceServer
}
