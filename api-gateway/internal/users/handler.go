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
	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind reg request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	w.WriteHeader(http.StatusOK)
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

func (uc *UsersClient) ExtJWTData(tokenString string) (UserInfo, error) {
	res, err := uc.client.ExtJWTData(context.Background(), &pb.ExtJWTDataReq{
		Token: tokenString,
	})
	if err != nil {
		uc.log.Error("Rpc request failed", zap.Error(err))
		return UserInfo{}, err
	}

	return UserInfo{
		Role: res.Role,
		UserID: res.UserId,
	}, nil
}
