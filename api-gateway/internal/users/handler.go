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
		Name  string `json:"name"`
		Role  string `json:"role"`
		Email string `json:"email"`
		Pswd  string `json:"password"`
	}{}

	uc.log.Debug("New reg user request")

	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind reg request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uc.log.Debug("Extracted data for reg user",
		zap.String("username", req.Name+req.Email),
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
		zap.String("username", req.Name+req.Email),
		zap.String("user role", req.Role),
		zap.String("token", res.Token))

	c.JSON(http.StatusOK, map[string]string{
		"token": res.Token,
	})
}

func (uc *UsersClient) logUser(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	req := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Pswd  string `json:"password"`
	}{}

	uc.log.Debug("New login request")

	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uc.log.Debug("Extracted data for login user",
		zap.String("username", req.Name+req.Email))

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

	uc.log.Debug("Successfully login",
		zap.String("username", req.Name+req.Email),
		zap.String("token", res.Token))

	c.JSON(http.StatusOK, map[string]string{
		"token": res.Token,
	})
}

func (uc *UsersClient) delUser(w http.ResponseWriter, r *http.Request) {
	req := struct {
		role string
		userId string
		delUserId string
	}{}
	req.role = r.Context().Value("userInfo").(UserInfo).Role
	req.userId = r.Context().Value("userInfo").(UserInfo).UserID
	req.delUserId = chi.URLParam(r, "delUserId")
	if req.delUserId == "me" {
		req.delUserId = req.userId
	}

	uc.log.Debug("New del user request",
		zap.String("user id", req.userId),
		zap.String("deleting user role", req.role),
		zap.String("deleted user id", req.delUserId))

	_, err := uc.client.DelUser(context.Background(), &pb.DelUserReq{
		Role: req.role,
		UserId: req.userId,
		DelUserId: req.delUserId,
	})
	if err != nil {
		uc.log.Error("Rpc request failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uc.log.Debug("Successfully deleted user",
		zap.String("deleted user id", req.delUserId))

	w.WriteHeader(http.StatusOK)
}

func (uc *UsersClient) ExtJWTData(tokenString string) (UserInfo, error) {
	res, err := uc.client.ExtJWTData(context.Background(), &pb.ExtJWTDataReq{
		Token: tokenString,
	})

	uc.log.Debug("New request in need jwt data",
		zap.String("token", tokenString))

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
	
	data, err := uc.ExtJWTData(tokenString)
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

