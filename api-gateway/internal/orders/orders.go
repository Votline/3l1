package orders

import (
	"os"

	"go.uber.org/zap"
	"github.com/go-chi/chi"
	"google.golang.org/grpc"

	"gateway/internal/service"
	pb "github.com/Votline/3l1/protos/generated-order"
)

type ordersClient struct {
	log *zap.Logger
	name string
	conn *grpc.ClientConn
	client pb.OrderServiceClient
}
func New(log *zap.Logger) service.Service {
	conn, err := grpc.NewClient(
		os.Getenv("OS_HOST")+":"+os.Getenv("OS_PORT"),
		grpc.WithInsecure())
	if err != nil {
		log.Fatal("Order-service connection failed")
	}
	return &ordersClient{
		log: log,
		conn: conn,
		name: "orders",
		client: pb.NewOrderServiceClient(conn),
	}
}

func (os *ordersClient) RegisterRoutes(g chi.Router) {
	g.Post("/", os.addOrder)
	g.Get("/{orderID}", os.orderInfo)
	g.Delete("/del/{orderID}", os.delOrder)
}

func (os *ordersClient) GetName() string {
	return os.name
}

func (os *ordersClient) Close() error {
	return os.conn.Close()
}
