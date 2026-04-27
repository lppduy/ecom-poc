package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OK(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, data)
}

func Created(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusCreated, data)
}

func BadRequest(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func NotFound(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusNotFound, gin.H{"error": message})
}

func InternalError(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
