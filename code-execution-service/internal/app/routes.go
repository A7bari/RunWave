package app

import (
	"errors"
	"net/http"

	"github.com/A7bari/RunWave/internal"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, podManager *internal.PodManager) {
	// Code execution handler
	router.POST("/execute", func(c *gin.Context) {
		var req internal.CodeExecutionReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		output, err := podManager.ExecuteCode(req.Code, req.Language)

		resp := &internal.CodeExecutionResp{
			Output:     output,
			StatusCode: http.StatusOK,
		}

		if err != nil {
			handleErrors(err, resp)
		}

		c.JSON(resp.StatusCode, resp)
	})

	// Health check handler
	router.GET("/health", func(c *gin.Context) {
		data := podManager.HealthzCheck()
		c.JSON(http.StatusOK, data)
	})
}

// ErrorHandler handles different types of errors and sends appropriate HTTP responses
func handleErrors(err error, resp *internal.CodeExecutionResp) {
	var ierr *internal.Error
	resp.Output = ""

	if errors.As(err, &ierr) {
		switch ierr.Code() {
		// Code execution errors
		case internal.ErrorCodeExecutionErr:
			resp.Error = err.Error()
			return
		case internal.ErrorCodeTimeout:
			resp.Error = "Execution timeout!"

		// Internal errors
		case internal.ErrorCodePodNotFound:
			resp.StatusCode = http.StatusServiceUnavailable
			resp.Error = "The service is currently unavailable, try again later!"
		case internal.ErrorCodeUnsupportLanguage:
			resp.StatusCode = http.StatusBadRequest
			resp.Error = "Unsupported language!"
		default:
			resp.StatusCode = http.StatusInternalServerError
			resp.Error = "Internal server error"
		}
	} else {
		resp.StatusCode = http.StatusInternalServerError
		resp.Error = "Internal server error"
	}
}
