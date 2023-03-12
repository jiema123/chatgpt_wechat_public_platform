package main

import (
	"github.com/ArtisanCloud/PowerLibs/v3/fmt"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/contract"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	models2 "github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount/server/handlers/models"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"myapp/config"
	"myapp/consts"
	handlers "myapp/handler"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var db = make(map[string]string)
var c = cache.New(5*time.Minute, 10*time.Minute)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.GET("check", weixinCheck)

	r.GET("wechat/callback", wechatCallback)

	r.POST("wechat/callback", wechatNotify)

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}

func wechatNotify(context *gin.Context) {
	conf := config.LoadConfig()
	OfficialAccountApp, err := officialAccount.NewOfficialAccount(&officialAccount.UserConfig{
		AppID:  conf.AppId,
		Secret: conf.SecretKey,

		Token:  conf.Token,
		AESKey: conf.AesKey,

		ResponseType: os.Getenv("response_type"),
		Log: officialAccount.Log{
			Level: "debug",
			File:  "./wechat.log",
		},
		HttpDebug: true,
		Debug:     false,
	})
	rs, err := OfficialAccountApp.Server.Notify(context.Request, func(event contract.EventInterface) interface{} {
		fmt.Dump("event", event)
		//return  "handle callback"

		switch event.GetMsgType() {
		case models2.CALLBACK_MSG_TYPE_TEXT:
			msg := models.MessageText{}
			err := event.ReadMessage(&msg)
			if err != nil {
				println(err.Error())
				return "error"
			}
			fmt.Dump(msg)

			// 查询锁，存在返回处理中的提示，避免重复操作
			_, ok := c.Get(consts.WECHAT_SESSION_KEY_LOCK + msg.FromUserName)
			if ok {
				return messages.NewText(consts.WECHAT_WAITE)
			}

			result, ok := c.Get(msg.FromUserName)
			if ok {
				c.Delete(msg.FromUserName)
				return messages.NewText(result.(string))
			}
			go handleMsg(msg.FromUserName, msg.Content)

			// 如果5秒内处理完那么直接返回
			time.Sleep(4 * time.Second)
			result2, ok2 := c.Get(msg.FromUserName)
			if ok2 {
				c.Delete(msg.FromUserName)
				return messages.NewText(result2.(string))
			}

			return messages.NewText(consts.WECHAT_HELLO)
		}
		//return messages.NewText("not supper")
		return kernel.SUCCESS_EMPTY_RESPONSE

	})
	if err != nil {
		panic(err)
	}

	text, _ := ioutil.ReadAll(rs.Body)
	context.String(http.StatusOK, string(text))

	if err != nil {
		panic(err)
	}
}

func handleMsg(userId string, msg string) {
	handlers.NewUserMessageHandler(msg, userId, c)
}

func wechatCallback(context *gin.Context) {
	conf := config.LoadConfig()
	OfficialAccountApp, err := officialAccount.NewOfficialAccount(&officialAccount.UserConfig{
		AppID:  conf.AppId,
		Secret: conf.SecretKey,

		Token:  conf.Token,
		AESKey: conf.AesKey,

		ResponseType: os.Getenv("response_type"),
		Log: officialAccount.Log{
			Level: "debug",
			File:  "./wechat.log",
		},
		HttpDebug: true,
		Debug:     false,
	})

	rs, err := OfficialAccountApp.Server.VerifyURL(context.Request)
	if err != nil {
		panic(err)
	}

	text, _ := ioutil.ReadAll(rs.Body)
	context.String(http.StatusOK, string(text))
}

func weixinCheck(ctx *gin.Context) {
	conf := config.LoadConfig()
	OfficialAccountApp, _ := officialAccount.NewOfficialAccount(&officialAccount.UserConfig{
		AppID:  conf.AppId,
		Secret: conf.SecretKey,

		Log: officialAccount.Log{
			Level: "debug",
			File:  "./wechat.log",
		},

		HttpDebug: true,
		Debug:     false,
	})
	r, _ := OfficialAccountApp.Base.GetCallbackIP(ctx)
	ctx.JSON(http.StatusOK, r)
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
