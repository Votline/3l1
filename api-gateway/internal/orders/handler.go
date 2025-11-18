package orders

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"gateway/internal/service"
	"gateway/internal/users"

	pb "github.com/Votline/3l1/protos/generated-order"
)

func (os *ordersClient) addOrder(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	req := struct{
		userID string
		TargetURL string `json:"target_url"`
		ServiceURL string `json:"service_url"`
		OrderType string `json:"order_type"`
		Quantity int32 `json:"quantity"`
	}{}

	os.log.Debug("New add order request")

	if err := c.Bind(&req); err != nil {
		os.log.Error("Failed to bind add order req", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.userID = r.Context().Value("userInfo").(users.UserInfo).UserID

	os.log.Debug("Extracted data for add order",
		zap.String("user id", req.userID),
		zap.String("target url", req.TargetURL),
		zap.String("service url", req.ServiceURL),
		zap.Int32("quantity", req.Quantity))

	res, err := os.client.AddOrder(c.Context(), &pb.AddOrderReq{
		UserId: req.userID,
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

	os.log.Debug("Successfully added order",
		zap.String("user id", req.userID),
		zap.String("added order id", res.Id))

	c.JSON(http.StatusOK, map[string]string{
		"id": res.Id,
	})
}

func (os *ordersClient) orderInfo(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	id := chi.URLParam(r, "orderID")
	userID := r.Context().Value("userInfo").(users.UserInfo).UserID

	os.log.Debug("New order info request",
		zap.String("user id", userID),
		zap.String("order id", id))

	res, err := os.client.OrderInfo(c.Context(), &pb.OrderInfoReq{
		Id: id,
		UserId: userID,
	})
	if err != nil {
		os.log.Error("Rpc request failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	os.log.Debug("Successfully extracted order data",
		zap.String("user id", userID),
		zap.String("order id", id))

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

func (os *ordersClient) delOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "orderID")
	role := r.Context().Value("userInfo").(users.UserInfo).Role
	userID := r.Context().Value("userInfo").(users.UserInfo).UserID

	os.log.Debug("New delete order request",
		zap.String("user id", userID),
		zap.String("user role", role),
		zap.String("order id", id))

	if _, err := os.client.DelOrder(context.Background(), &pb.DelOrderReq{
		Id: id,
		Role: role,
		UserId: userID,
	}); err != nil {
		os.log.Error("Rpc request failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	os.log.Debug("Successfully deleted order",
		zap.String("user id", userID),
		zap.String("user role", role),
		zap.String("order id", id))

	w.WriteHeader(http.StatusOK)
}
