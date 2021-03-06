package api

import (
	"github.com/gin-gonic/gin"
	"github.com/saltbo/gopkg/ginutil"
	"github.com/saltbo/gopkg/jwtutil"
	"github.com/saltbo/gopkg/strutil"
	"github.com/spf13/viper"

	"github.com/saltbo/zpan/internal/app/dao"
	"github.com/saltbo/zpan/internal/app/model"
	"github.com/saltbo/zpan/internal/pkg/bind"
	"github.com/saltbo/zpan/internal/pkg/gormutil"
	"github.com/saltbo/zpan/internal/pkg/middleware"
	"github.com/saltbo/zpan/internal/pkg/provider"
)

type Option struct {
	jwtutil.JWTUtil
}

func NewOptionResource() *Option {
	return &Option{}
}

func (rs *Option) Register(router *gin.RouterGroup) {
	router.PUT("/system/database", rs.createDatabase)
	router.PUT("/system/account", rs.createAdministrator)

	router.Use(middleware.Installer)
	router.GET("/system/providers", rs.providers)
	router.GET("/system/options/:name", rs.find)
	router.PUT("/system/options/:name", rs.update)
}

func (rs *Option) createDatabase(c *gin.Context) {
	p := make(map[string]string)
	if err := c.ShouldBind(&p); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	if _, err := gormutil.New(gormutil.Config{Driver: p["driver"], DSN: p["dsn"]}); err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	viper.Set("database.driver", p["driver"])
	viper.Set("database.dsn", p["dsn"])
	if err := viper.WriteConfigAs("config.yml"); err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	ginutil.JSON(c)
}

func (rs *Option) createAdministrator(c *gin.Context) {
	p := new(bind.BodyUserCreation)
	if err := c.ShouldBind(&p); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}
	// 创建基本信息
	user := &model.User{
		Email:    p.Email,
		Username: "admin",
		Password: strutil.Md5Hex(p.Password),
		Roles:    "admin",
		Ticket:   strutil.RandomText(6),
		Status:   model.StatusActivated,
	}
	if _, err := dao.NewUser().Create(user, 0); err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	ginutil.JSON(c)

}

func (rs *Option) providers(c *gin.Context) {
	ginutil.JSONData(c, provider.GetProviders())
}

func (rs *Option) find(c *gin.Context) {
	ret, err := dao.NewOption().Get(c.Param("name"))
	if err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	ginutil.JSONData(c, ret)
}

func (rs *Option) update(c *gin.Context) {
	p := make(map[string]interface{})
	if err := c.ShouldBind(&p); err != nil {
		ginutil.JSONBadRequest(c, err)
		return
	}

	if err := dao.NewOption().Set(c.Param("name"), p); err != nil {
		ginutil.JSONServerError(c, err)
		return
	}

	ginutil.JSON(c)
}
