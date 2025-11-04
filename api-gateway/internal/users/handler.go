package users

import (
	"net/http"

	"go.uber.org/zap"

	"gateway/internal/service"
	pb "github.com/Votline/3l1/protos/generated-user"
)

func (uc usersClient) regUser(w http.ResponseWriter, r *http.Request) {
	c := service.NewContext(w, r)
	req := struct{
		name string `json:"name"`
		role string `json:"role"`
		pswd string `json:"password"`
	}{}
	if err := c.Bind(&req); err != nil {
		uc.log.Error("Failed to bind reg request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashRes, err := uc.client.HashPswd(c.Context(), &pb.HashReq{
		Password: req.pswd,
	})
	if err != nil {
		uc.log.Error("Failed to hash password", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := uc.client.RegUser(c.Context(), &pb.RegReq{
		Name: req.name,
		Role: req.role,
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
