package utils

import "github.com/gin-gonic/gin"

// SuccessResponse returns a standardized success response
func SuccessResponse(message string, data interface{}) gin.H {
	return gin.H{
		"success": true,
		"message": message,
		"data":    data,
	}
}

// ErrorResponse returns a standardized error response
func ErrorResponse(message string) gin.H {
	return gin.H{
		"success": false,
		"error":   message,
	}
}
