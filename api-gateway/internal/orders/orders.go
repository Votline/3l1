package orders

import (
	"context"
	"os"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	gc "gateway/internal/graceful"
	"gateway/internal/service"
	pb "github.com/Votline/3l1/protos/generated-order"
)

type ordersClient struct {
	log     *zap.Logger
	name    string
	conn    *grpc.ClientConn
	client  pb.OrderServiceClient
	hist    *prometheus.HistogramVec
	counter *prometheus.CounterVec
	active  prometheus.Gauge
}

func New(resTime *prometheus.HistogramVec, log *zap.Logger) service.Service {
	conn, err := grpc.NewClient(
		os.Getenv("OS_HOST")+":"+os.Getenv("OS_PORT"),
		grpc.WithInsecure())
	if err != nil {
		log.Fatal("Order-service connection failed")
	}
	return &ordersClient{
		log:     log,
		conn:    conn,
		name:    "orders",
		client:  pb.NewOrderServiceClient(conn),
		hist:    resTime,
		counter: newCounter(),
		active:  newGauge(),
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

func (os *ordersClient) Close(ctx context.Context) error {
	return gc.Shutdown(os.conn.Close, ctx)
}

func (os *ordersClient) NewTimer(svc, oper string) *prometheus.Timer {
	return prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		os.hist.WithLabelValues(svc, oper).Observe(v)
	}))
}

func (os *ordersClient) GetCounter(label string) prometheus.Counter {
	return os.counter.WithLabelValues(label)
}

func (os *ordersClient) GetActive() prometheus.Gauge {
	return os.active
}

func newCounter() *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orders_operation_total",
		Help: "Total number of operations for order service",
	}, []string{"operation"})
}

func newGauge() prometheus.Gauge {
	return promauto.NewGauge(prometheus.GaugeOpts{
		Name: "orders_active_operations",
		Help: "Total number of active operations for order service",
	})
}
