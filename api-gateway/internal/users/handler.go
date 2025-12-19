package users

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	ck "gateway/internal/contextKeys"
	"gateway/internal/service"
	pb "github.com/Votline/3l1/protos/generated-user"
)

func (uc *UsersClient) regUser(w http.ResponseWriter, r *http.Request) {
	const op = "usersClient.regUser"

	c := service.NewContext(w, r)
	req := struct {
		Name  string `json:"name"     validate:"required,min=2,max=50"`
		Role  string `json:"role"     validate:"oneof=admin user guest dev"`
		Email string `json:"email"    validate:"email"`
		Pswd  string `json:"password" validate:"required,min=8"`
	}{}

	rq := r.Context().Value(ck.ReqKey).(string)
	uc.log.Info("New request",
		zap.String("op", op),
		zap.String("request id", rq))

	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind reg request",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.Validate(req); err != nil {
		uc.log.Error("Failed to validate request data",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		return
	}

	uc.log.Info("Extracted data for reg user",
		zap.String("role", req.Role))

	res, err := service.Execute(uc.cb, func() (*pb.RegRes, error) {
		return uc.client.RegUser(c.Context(), &pb.RegReq{
			Name:      req.Name,
			Email:     req.Email,
			Role:      req.Role,
			Password:  req.Pswd,
			RequestId: rq,
		})
	})

	if err != nil {
		uc.log.Error("Rpc request failed",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uc.log.Info("Successfully added user",
		zap.String("user role", req.Role))

	c.SetSession(res.SessionKey)

	c.JSON(http.StatusOK, map[string]string{
		"token": res.Token,
	})
}

func (uc *UsersClient) logUser(w http.ResponseWriter, r *http.Request) {
	const op = "usersClient.logUser"

	c := service.NewContext(w, r)
	req := struct {
		Name  string `json:"name"     validate:"required,min=2,max=50"`
		Email string `json:"email"    validate:"email"`
		Pswd  string `json:"password" validate:"required,min=8"`
	}{}

	rq := r.Context().Value(ck.ReqKey).(string)
	uc.log.Info("New request",
		zap.String("op", op),
		zap.String("request id", rq))

	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind request",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.Validate(req); err != nil {
		uc.log.Error("Failed to validate request data",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		return
	}

	uc.log.Info("Successfully extract data")

	res, err := service.Execute(uc.cb, func() (*pb.LogRes, error) {
		return uc.client.LogUser(c.Context(), &pb.LogReq{
			Name:      req.Name,
			Email:     req.Email,
			Password:  req.Pswd,
			RequestId: rq,
		})
	})
	if err != nil {
		uc.log.Error("Rpc request failed",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uc.log.Info("Successfully login")

	c.SetSession(res.SessionKey)
	c.JSON(http.StatusOK, map[string]string{
		"token": res.Token,
	})
}

func (uc *UsersClient) delUser(w http.ResponseWriter, r *http.Request) {
	const op = "usersClient.delUser"

	c := service.NewContext(w, r)
	req := struct {
		sk        string `validate:"required,len=36"`
		role      string `validate:"oneof=admin user guest dev"`
		userId    string `validate:"required,len=36"`
		delUserId string `validate:"required,len=36"`
	}{}

	rq := r.Context().Value(ck.ReqKey).(string)

	req.role = r.Context().Value(ck.UserKey).(ck.UserInfo).Role
	req.userId = r.Context().Value(ck.UserKey).(ck.UserInfo).UserID
	req.delUserId = chi.URLParam(r, "delUserId")
	if req.delUserId == "me" {
		req.delUserId = req.userId
	}
	sk, err := r.Cookie("session_key")
	if err != nil {
		uc.log.Error("Couldn't get session key from cookies",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		return
	}
	req.sk = sk.Value

	if err := c.Validate(req); err != nil {
		uc.log.Error("Failed to validate request data",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		return
	}

	uc.log.Info("New request",
		zap.String("op", op),
		zap.String("request id", rq))

	if _, err := service.Execute(uc.cb, func() (*pb.DelUserRes, error) {
		return uc.client.DelUser(context.Background(), &pb.DelUserReq{
			Role:       req.role,
			UserId:     req.userId,
			DelUserId:  req.delUserId,
			SessionKey: req.sk,
			RequestId:  rq,
		})
	}); err != nil {
		uc.log.Error("Rpc request failed",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uc.log.Info("Successfully deleted user",
		zap.String("deleted user id", req.delUserId))

	w.WriteHeader(http.StatusOK)
}

func (uc *UsersClient) ExtJWTData(tokenString, sk, rq string) (ck.UserInfo, error) {
	const op = "usersClient.ExtJWTData"

	c := service.NewContext(nil, nil)
	req := struct {
		token string `validator:"required,min=100"`
		sk    string `validator:"required,min=36"`
	}{}

	req.token, req.sk = tokenString, sk

	if err := c.Validate(req); err != nil {
		uc.log.Error("Failed to validate request data",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		return ck.UserInfo{}, err
	}

	uc.log.Info("New request",
		zap.String("op", op),
		zap.String("request id", rq))

	res, err := service.Execute(uc.cb, func() (*pb.ExtJWTDataRes, error) {
		return uc.client.ExtJWTData(context.Background(), &pb.ExtJWTDataReq{
			Token:      req.token,
			SessionKey: req.sk,
			RequestId:  rq,
		})
	})

	if err != nil {
		uc.log.Error("Rpc request failed",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		return ck.UserInfo{}, err
	}

	uc.log.Info("Successfully extracted data from jwt token",
		zap.String("user id", res.UserId),
		zap.String("role", res.Role))

	return ck.UserInfo{
		Role:   res.Role,
		UserID: res.UserId,
	}, nil
}

func (uc *UsersClient) extUserId(w http.ResponseWriter, r *http.Request) {
	const op = "usersClient.extUserId"

	rq := r.Context().Value(ck.ReqKey).(string)
	uc.log.Info("New request",
		zap.String("op", op),
		zap.String("request id", rq))

	tokenString := chi.URLParam(r, "token")
	sk, err := r.Cookie("session_key")
	if err != nil {
		uc.log.Error("Couldn't get session key from cookies",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := uc.ExtJWTData(tokenString, sk.Value, rq)
	if err != nil {
		uc.log.Error("Failed to extract jwt data",
			zap.String("op", op),
			zap.String("request id", rq),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := service.NewContext(w, r)
	c.JSON(http.StatusOK, map[string]string{
		"user_id": data.UserID,
	})
}
