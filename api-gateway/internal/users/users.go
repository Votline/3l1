package users

import (
	"os"

	"go.uber.org/zap"
	"github.com/go-chi/chi"
	"google.golang.org/grpc"

	"gateway/internal/service"

	pb "github.com/Votline/3l1/protos/generated-user"
)

type usersClient struct {
	log *zap.Logger
	name string
	conn *grpc.ClientConn
	client pb.UserServiceClient
}

func New(log *zap.Logger) service.Service {
	conn, err := grpc.NewClient(
		os.Getenv("US_HOST")+":"+os.Getenv("US_PORT"),
		grpc.WithInsecure())
	if err != nil {
		log.Fatal("User-service connection failed", zap.Error(err))
	}
	return &usersClient{
		log: log,
		conn: conn,
		name: "users",
		client: pb.NewUserServiceClient(conn),
	}
}

func (uc *usersClient) RegisterRoutes(g chi.Router) {
	g.Post("/reg", uc.regUser)
	g.Post("/log", uc.logUser)
}

func (uc *usersClient) GetName() string {
	return uc.name
}

func (uc *usersClient) Close() error {
	return uc.conn.Close()
}
