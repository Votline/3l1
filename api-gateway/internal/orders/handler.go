package orders

import (
	"net/http"

	"gateway/internal/service"

	pb "github.com/Votline/3l1/protos/generated-order"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func (os *ordersClient) addOrder(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	req := struct{
		UserID string `json:"user_id"`
		TargetURL string `json:"target_url"`
		ServiceURL string `json:"service_url"`
		OrderType string `json:"order_type"`
		Quantity int32 `json:"quantity"`
	}{}
	if err := c.Bind(&req); err != nil {
		os.log.Error("Failed to bind add order req", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := os.client.AddOrder(c.Context(), &pb.AddOrderReq{
		UserId: req.UserID,
		TargetUrl: req.TargetURL,
		ServiceUrl: req.ServiceURL,
		OrderType: req.OrderType,
		Quantity: req.Quantity,
	})
	if err != nil {
		os.log.Error("Rpc request failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"id": res.Id,
	})
}

func (os ordersClient) orderInfo(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	id := chi.URLParam(r, "orderID")

	res, err := os.client.OrderInfo(c.Context(), &pb.OrderInfoReq{
		Id: id,
	})
	if err != nil {
		os.log.Error("Rpc request failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"user_id": res.UserId,
		"status": res.Status,
		"target_url": res.TargetUrl,
		"service_url": res.ServiceUrl,
		"order_type": res.OrderType,
		"created_at": res.CreatedAt.String(),
		"updated_at": res.UpdatedAt.String(),
	})
}
