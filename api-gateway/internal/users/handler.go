package users

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"github.com/go-chi/chi"

	"gateway/internal/service"
	pb "github.com/Votline/3l1/protos/generated-user"
)

type UserInfo struct {
	Role string
	UserID string
}

func (uc *UsersClient) regUser(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	req := struct{
		Name  string `json:"name"     validate:"required,min=2,max=50"`
		Role  string `json:"role"     validate:"oneof=admin user guest dev"`
		Email string `json:"email"    validate:"email"`
		Pswd  string `json:"password" validate:"required,min=8"`
	}{}

	uc.log.Debug("New reg user request")

	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind reg request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.Validate(req); err != nil {
		uc.log.Error("Failed to validate request data", zap.Error(err))
		return
	}

	uc.log.Debug("Extracted data for reg user",
		zap.String("role", req.Role))

	hashRes, err := uc.client.HashPswd(c.Context(), &pb.HashReq{
		Password: req.Pswd,
	})
	if err != nil {
		uc.log.Error("Failed to hash password", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := uc.client.RegUser(c.Context(), &pb.RegReq{
		Name:  req.Name,
		Email: req.Email,
		Role:  req.Role,
		PasswordHash: hashRes.PasswordHash,
	})
	if err != nil {
		uc.log.Error("Rpc request failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uc.log.Debug("Successfully added user",
		zap.String("user role", req.Role))

	c.SetSession(res.SessionKey)
	c.JSON(http.StatusOK, map[string]string{
		"token": res.Token,
	})
}

func (uc *UsersClient) logUser(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	req := struct {
		Name  string `json:"name"     validator:"required,min=2,max=50"`
		Email string `json:"email"    validator:"email"`
		Pswd  string `json:"password" validator:"required,min=8"`
	}{}

	uc.log.Debug("New login request")

	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.Validate(req); err != nil {
		uc.log.Error("Failed to validate request data", zap.Error(err))
		return
	}

	uc.log.Debug("Successfully extract data")

	res, err := uc.client.LogUser(c.Context(), &pb.LogReq{
		Name: req.Name,
		Email: req.Email,
		Password: req.Pswd,
	})
	if err != nil {
		uc.log.Error("Rpc request failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uc.log.Debug("Successfully login")

	c.SetSession(res.SessionKey)
	c.JSON(http.StatusOK, map[string]string{
		"token": res.Token,
	})
}

func (uc *UsersClient) delUser(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	req := struct {
		sk string `validator:"required,len=36"`
		role string `validator:"oneof=admin user guest dev"`
		userId string `validtor:"required,len=36"`
		delUserId string `validtor:"required,len=36"`
	}{}

	req.role = r.Context().Value("userInfo").(UserInfo).Role
	req.userId = r.Context().Value("userInfo").(UserInfo).UserID
	req.delUserId = chi.URLParam(r, "delUserId")
	if req.delUserId == "me" {
		req.delUserId = req.userId
	}
	sk, err := r.Cookie("session_key")
	if err != nil {
		uc.log.Error("Couldn't get session key from cookies", zap.Error(err))
		return
	}
	req.sk = sk.Value

	if err := c.Validate(req); err != nil {
		uc.log.Error("Failed to validate request data", zap.Error(err))
		return
	}

	uc.log.Debug("New del user request",
		zap.String("user id", req.userId),
		zap.String("deleting user role", req.role),
		zap.String("deleted user id", req.delUserId))

	if _, err := uc.client.DelUser(context.Background(), &pb.DelUserReq{
		Role: req.role,
		UserId: req.userId,
		DelUserId: req.delUserId,
		SessionKey: req.sk,
	}); err != nil {
		uc.log.Error("Rpc request failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uc.log.Debug("Successfully deleted user",
		zap.String("deleted user id", req.delUserId))

	w.WriteHeader(http.StatusOK)
}

func (uc *UsersClient) ExtJWTData(tokenString, sk string) (UserInfo, error) {
	c := service.NewContext(nil, nil)
	req := struct{
		token string `validator:"required,min=100"`
		sk string `validator:"required,min=36"`
	}{}

	req.token, req.sk = tokenString, sk

	if err := c.Validate(req); err != nil {
		uc.log.Error("Failed to validate request data", zap.Error(err))
		return UserInfo{}, err
	}

	res, err := uc.client.ExtJWTData(context.Background(), &pb.ExtJWTDataReq{
		Token: req.token,
		SessionKey: req.sk,
	})

	uc.log.Debug("New request in need jwt data")

	if err != nil {
		uc.log.Error("Rpc request failed", zap.Error(err))
		return UserInfo{}, err
	}

	uc.log.Debug("Successfully extracted data from jwt token",
		zap.String("user id", res.UserId),
		zap.String("role", res.Role))

	return UserInfo{
		Role: res.Role,
		UserID: res.UserId,
	}, nil
}

func (uc *UsersClient) extUserId(w http.ResponseWriter, r *http.Request) {
	tokenString := chi.URLParam(r, "token")
	sk, err := r.Cookie("session_key")
	if err != nil {
		uc.log.Error("Couldn't get session key from cookies", zap.Error(err))
	}

	data, err := uc.ExtJWTData(tokenString, sk.Value)
	if err != nil {
		uc.log.Error("Failed to extract jwt data", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	c := service.NewContext(w, r)
	c.JSON(http.StatusOK, map[string]string{
		"user_id": data.UserID,
	})
}
