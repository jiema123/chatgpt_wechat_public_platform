package handlers

import (
	"github.com/patrickmn/go-cache"
	"myapp/config"
	"time"
)

const deadlineExceededText = "你的提问暂时无法得到回复，请重新提问"

var c = cache.New(config.LoadConfig().SessionTimeout, time.Minute*5)

// MessageHandlerInterface 消息处理接口
type MessageHandlerInterface interface {
	handle() error
	ReplyText() error
}
