package users

import (
	"os"

	"go.uber.org/zap"
	"github.com/go-chi/chi"
	"google.golang.org/grpc"

	"gateway/internal/service"

	pb "github.com/Votline/3l1/protos/generated-user"
)

type UsersClient struct {
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
	return &UsersClient{
		log: log,
		conn: conn,
		name: "users",
		client: pb.NewUserServiceClient(conn),
	}
}

func (uc *UsersClient) RegisterRoutes(g chi.Router) {
	g.Post("/reg", uc.regUser)
	g.Post("/log", uc.logUser)
	g.Delete("/del/{delUserId}", uc.delUser)
	g.Get("/extUserId/{token}", uc.extUserId)
}

func (uc *UsersClient) GetName() string {
	return uc.name
}

func (uc *UsersClient) Close() error {
	return uc.conn.Close()
}
