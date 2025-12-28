package route

import (
	"gin_base/app/controller/common"
	"gin_base/app/helper/response_helper"
	"gin_base/app/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(e *gin.Engine) {
	e.NoRoute(func(context *gin.Context) {
		response_helper.Common(context, 404, "路由不存在")
	})

	// 邮件记录首页
	e.GET("/", common.EmailLogIndex)

	// favicon
	e.StaticFile("/favicon.png", "./static/image/favicon.png")

	api := e.Group("/api")
	api.GET("/test", common.Test)
	api.Any("/email", common.Email)

	// 邮件记录API
	api.POST("/getEmailLogList", common.GetEmailLogList)
	api.POST("/deleteEmailLog", common.DeleteEmailLog)

	//登录相关
	auth := api.Group("", middleware.Auth())
	auth.POST("/test_auth", common.Test)
}
