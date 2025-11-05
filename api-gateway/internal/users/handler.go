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

func (uc usersClient) logUser(w http.ResponseWriter, r *http.Request) {
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
		PasswordHash: req.Pswd,
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
