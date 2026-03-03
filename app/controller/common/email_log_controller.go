package common

import (
	"gin_base/app/helper/db_helper"
	"gin_base/app/helper/exception_helper"
	"gin_base/app/helper/request_helper"
	"gin_base/app/helper/response_helper"
	"gin_base/app/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"os"
	"time"
)

// EmailLogIndex 邮件记录首页
func EmailLogIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "email_log.html", nil)
}

// GetEmailLogList 邮件记录列表API
func GetEmailLogList(c *gin.Context) {
	type Param struct {
		AuthCode  string `json:"auth_code" mapstructure:"auth_code" validate:"required" label:"授权码"`
		Keyword   string `json:"keyword" mapstructure:"keyword" validate:"omitempty" label:"关键词"`
		StartDate string `json:"start_date" mapstructure:"start_date" validate:"omitempty" label:"开始日期"`
		EndDate   string `json:"end_date" mapstructure:"end_date" validate:"omitempty" label:"结束日期"`
		Success   string `json:"success" mapstructure:"success" validate:"omitempty" label:"发送状态"`
	}
	var param Param
	request_helper.InputStruct(c, &param)

	// 验证授权码（页面专用）
	if param.AuthCode != os.Getenv("WEB_AUTH_CODE") {
		exception_helper.CommonException("授权码错误")
	}

	// 基础查询条件（日期+关键词）
	baseQuery := func(db *gorm.DB) *gorm.DB {
		// 开始日期筛选
		if param.StartDate != "" {
			startTime, _ := time.ParseInLocation("2006-01-02", param.StartDate, time.Local)
			db = db.Where("created_at >= ?", startTime)
		}
		// 结束日期筛选
		if param.EndDate != "" {
			endTime, _ := time.ParseInLocation("2006-01-02", param.EndDate, time.Local)
			endTime = endTime.Add(24*time.Hour - time.Second) // 结束日期当天 23:59:59
			db = db.Where("created_at <= ?", endTime)
		}
		// 关键词模糊查询
		if param.Keyword != "" {
			keyword := "%" + param.Keyword + "%"
			db = db.Where("to_email LIKE ? OR subject LIKE ? OR body LIKE ? OR request_ip LIKE ?",
				keyword, keyword, keyword, keyword)
		}
		return db
	}

	// 统计成功数量
	var successCount int64
	baseQuery(db_helper.Db().Model(&model.EmailLog{})).Where("success = 1").Count(&successCount)

	// 统计失败数量
	var failedCount int64
	baseQuery(db_helper.Db().Model(&model.EmailLog{})).Where("success = 0").Count(&failedCount)

	// 构建列表查询
	db := baseQuery(db_helper.Db().Model(&model.EmailLog{})).Order("id DESC")

	// 发送状态筛选
	if param.Success == "1" {
		db = db.Where("success = 1")
	} else if param.Success == "0" {
		db = db.Where("success = 0")
	}

	// 分页查询
	result := db_helper.AutoPage(c, db)
	result["success_count"] = successCount
	result["failed_count"] = failedCount
	response_helper.Success(c, "查询成功", result)
}

// DeleteEmailLog 删除筛选结果
func DeleteEmailLog(c *gin.Context) {
	type Param struct {
		AuthCode  string `json:"auth_code" mapstructure:"auth_code" validate:"required" label:"授权码"`
		Keyword   string `json:"keyword" mapstructure:"keyword" validate:"omitempty" label:"关键词"`
		StartDate string `json:"start_date" mapstructure:"start_date" validate:"omitempty" label:"开始日期"`
		EndDate   string `json:"end_date" mapstructure:"end_date" validate:"omitempty" label:"结束日期"`
		Success   string `json:"success" mapstructure:"success" validate:"omitempty" label:"发送状态"`
	}
	var param Param
	request_helper.InputStruct(c, &param)

	// 验证授权码（页面专用）
	if param.AuthCode != os.Getenv("WEB_AUTH_CODE") {
		exception_helper.CommonException("授权码错误")
	}

	// 构建删除条件
	db := db_helper.Db().Model(&model.EmailLog{})

	// 开始日期筛选
	if param.StartDate != "" {
		startTime, _ := time.ParseInLocation("2006-01-02", param.StartDate, time.Local)
		db = db.Where("created_at >= ?", startTime)
	}
	// 结束日期筛选
	if param.EndDate != "" {
		endTime, _ := time.ParseInLocation("2006-01-02", param.EndDate, time.Local)
		endTime = endTime.Add(24*time.Hour - time.Second)
		db = db.Where("created_at <= ?", endTime)
	}
	// 关键词模糊查询
	if param.Keyword != "" {
		keyword := "%" + param.Keyword + "%"
		db = db.Where("to_email LIKE ? OR subject LIKE ? OR body LIKE ? OR request_ip LIKE ?",
			keyword, keyword, keyword, keyword)
	}
	// 发送状态筛选
	if param.Success == "1" {
		db = db.Where("success = 1")
	} else if param.Success == "0" {
		db = db.Where("success = 0")
	}

	// 执行删除
	result := db.Delete(&model.EmailLog{})
	if result.Error != nil {
		exception_helper.CommonException("删除失败: " + result.Error.Error())
	}

	response_helper.Success(c, "删除成功", map[string]interface{}{
		"deleted_count": result.RowsAffected,
	})
}
