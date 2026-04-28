package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/services/auth/internal/api/response"
	"github.com/lppduy/ecom-poc/services/auth/internal/domain"
	"github.com/lppduy/ecom-poc/services/auth/internal/service"
)

type AuthController struct {
	svc       service.AuthService
	jwtSecret string
}

func NewAuthController(svc service.AuthService, jwtSecret string) *AuthController {
	return &AuthController{svc: svc, jwtSecret: jwtSecret}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := ctrl.svc.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUsernameTaken) {
			response.Conflict(c, "username already taken")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, gin.H{"id": user.ID, "username": user.Username, "created_at": user.CreatedAt})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	token, err := ctrl.svc.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPassword) {
			response.Unauthorized(c, "invalid username or password")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token})
}

func (ctrl *AuthController) Me(c *gin.Context) {
	userID := jwtutil.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "unauthenticated")
		return
	}
	user, err := ctrl.svc.Me(userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.Unauthorized(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": user.ID, "username": user.Username, "created_at": user.CreatedAt})
}
