package orders

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	ck "gateway/internal/contextKeys"
	"gateway/internal/service"

	pb "github.com/Votline/3l1/protos/generated-order"
)

func (oc *ordersClient) addOrder(w http.ResponseWriter, r *http.Request) {
	const op = "ordersClient.addOrder"

	c := service.NewContext(w, r)
	req := struct {
		userID     string `validate:"required,len=36"`
		TargetURL  string `json:"target_url" validate:"url"`
		ServiceURL string `json:"service_url" validate:"url"`
		OrderType  string `json:"order_type" validate:"oneof=comments likes views"`
		Quantity   int32  `json:"quantity" validate:"gt=1"`
	}{}

	rq := r.Context().Value(ck.ReqKey).(string)
	oc.log.Debug("New add order request",
		zap.String("op", op),
		zap.String("request id", rq))

	if err := c.Bind(&req); err != nil {
		oc.log.Error("Failed to bind add order req",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.userID = r.Context().Value("userInfo").(ck.UserInfo).UserID

	if err := c.Validate(req); err != nil {
		oc.log.Error("Failed to validate request data",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		return
	}

	oc.log.Debug("Extracted data for add order",
		zap.String("op", op),
		zap.String("request id", rq),
		zap.String("user id", req.userID),
		zap.String("target url", req.TargetURL),
		zap.String("service url", req.ServiceURL),
		zap.Int32("quantity", req.Quantity))

	res, err := service.Execute(oc.cb, func() (*pb.AddOrderRes, error) {
		return oc.client.AddOrder(c.Context(), &pb.AddOrderReq{
			UserId:     req.userID,
			TargetUrl:  req.TargetURL,
			ServiceUrl: req.ServiceURL,
			OrderType:  req.OrderType,
			Quantity:   req.Quantity,
			RequestId:  rq,
		})
	})
	if err != nil {
		oc.log.Error("Rpc request failed",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oc.log.Debug("Successfully added order",
		zap.String("op", op),
		zap.String("request id", rq),
		zap.String("user id", req.userID),
		zap.String("added order id", res.Id))

	c.JSON(http.StatusOK, map[string]string{
		"id": res.Id,
	})
}

func (oc *ordersClient) orderInfo(w http.ResponseWriter, r *http.Request) {
	const op = "ordersClient.orderInfo"

	c := service.NewContext(w, r)
	req := struct {
		id     string `validate:"required,len=36"`
		userID string `validate:"required,len=36"`
	}{}

	req.id = chi.URLParam(r, "orderID")
	req.userID = r.Context().Value("userInfo").(ck.UserInfo).UserID
	rq := r.Context().Value(ck.ReqKey).(string)

	if err := c.Validate(req); err != nil {
		oc.log.Error("Failed to validate request data",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		return
	}

	oc.log.Debug("New order info request",
		zap.String("op", op),
		zap.String("request id", rq),
		zap.String("user id", req.userID),
		zap.String("order id", req.id))

	res, err := service.Execute(oc.cb, func() (*pb.OrderInfoRes, error) {
		return oc.client.OrderInfo(c.Context(), &pb.OrderInfoReq{
			Id:        req.id,
			UserId:    req.userID,
			RequestId: rq,
		})
	})
	if err != nil {
		oc.log.Error("Rpc request failed",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oc.log.Debug("Successfully extracted order data",
		zap.String("op", op),
		zap.String("request id", rq),
		zap.String("user id", req.userID),
		zap.String("order id", req.id))

	c.JSON(http.StatusOK, map[string]string{
		"user_id":     res.UserId,
		"status":      res.Status,
		"target_url":  res.TargetUrl,
		"service_url": res.ServiceUrl,
		"order_type":  res.OrderType,
		"created_at":  res.CreatedAt.String(),
		"updated_at":  res.UpdatedAt.String(),
	})
}

func (oc *ordersClient) delOrder(w http.ResponseWriter, r *http.Request) {
	const op = "ordersClient.delOrder"

	c := service.NewContext(w, r)
	req := struct {
		id     string `validate:"required,len=36"`
		role   string `validate:"oneof=admin user guest dev"`
		userID string `validate:"required,len=36"`
	}{}

	req.id = chi.URLParam(r, "orderID")
	req.role = r.Context().Value("userInfo").(ck.UserInfo).Role
	req.userID = r.Context().Value("userInfo").(ck.UserInfo).UserID
	rq := r.Context().Value(ck.ReqKey).(string)

	if err := c.Validate(req); err != nil {
		oc.log.Error("Failed to validate request data",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oc.log.Debug("New delete order request",
		zap.String("op", op),
		zap.String("request id", rq),
		zap.String("user id", req.userID),
		zap.String("user role", req.role),
		zap.String("order id", req.id))

	if _, err := service.Execute(oc.cb, func() (*pb.DelOrderRes, error) {
		return oc.client.DelOrder(context.Background(), &pb.DelOrderReq{
			Id:        req.id,
			Role:      req.role,
			UserId:    req.userID,
			RequestId: rq,
		})
	}); err != nil {
		oc.log.Error("Rpc request failed",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oc.log.Debug("Successfully deleted order",
		zap.String("op", op),
		zap.String("request id", rq),
		zap.String("user id", req.userID),
		zap.String("user role", req.role),
		zap.String("order id", req.id))

	w.WriteHeader(http.StatusOK)
}
