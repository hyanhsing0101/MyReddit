package controller

import "github.com/gin-gonic/gin"

func DebugAuthAnyHandler(c *gin.Context) {
	ResponseSuccess(c, gin.H{"need": "login", "ok": true})
}

func DebugAuthAdminHandler(c *gin.Context) {
	ResponseSuccess(c, gin.H{"need": "site_admin", "ok": true})
}
