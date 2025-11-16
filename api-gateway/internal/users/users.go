package users

import (
	"os"

	"go.uber.org/zap"
	"github.com/go-chi/chi"
	"google.golang.org/grpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"gateway/internal/service"

	pb "github.com/Votline/3l1/protos/generated-user"
)

type UsersClient struct {
	log *zap.Logger
	name string
	conn *grpc.ClientConn
	client pb.UserServiceClient
	hist *prometheus.HistogramVec
	Counter *prometheus.CounterVec
	Active prometheus.Gauge
}

func New(resTime *prometheus.HistogramVec ,log *zap.Logger) service.Service {
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
		hist: resTime,
		Counter: newCounter(),
		Active: newGauge(),
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

func (uc *UsersClient) NewTimer(label string) *prometheus.Timer{
	return prometheus.NewTimer(prometheus.ObserverFunc(func(v float64){
		uc.hist.WithLabelValues(label).Observe(v)
	}))
}

func newCounter() *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "users_operation_total",
		Help: "Total number of operations for user service",
	}, []string{"users_service_operation"})
}

func newGauge() prometheus.Gauge {
	gauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "users_active_operations",
		Help: "Total number of active operations for user service",
	})
	return gauge
}
