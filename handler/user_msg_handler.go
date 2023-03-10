package handlers

import (
	"github.com/patrickmn/go-cache"
	"log"
	"myapp/config"
	"myapp/gpt"
	"myapp/service"
	"strings"
)

// UserMessageHandler 私聊消息处理
type UserMessageHandler struct {
	// 接收到消息
	msg string
	// 发送的用户
	userId string
	// 实现的用户业务
	service service.UserServiceInterface
}

// NewUserMessageHandler 创建私聊处理器
func NewUserMessageHandler(message string, userId string, cache *cache.Cache) (*UserMessageHandler, error) {
	userService := service.NewUserService(c, userId)
	handler := &UserMessageHandler{
		msg:     message,
		userId:  userId,
		service: userService,
	}
	cache.Set(userId, handler.ReplyText(), -1)
	return handler, nil
}

// ReplyText 发送文本消息到群
func (h *UserMessageHandler) ReplyText() string {
	log.Printf("Received User[%v], Content[%v]", h.userId, h.msg)
	var (
		reply string
		err   error
	)
	// 1.获取上下文，如果字符串为空不处理
	requestText := h.getRequestText()
	if requestText == "" {
		log.Println("user message is empty")
		return deadlineExceededText
	}

	// 2.向GPT发起请求，如果回复文本等于空,不回复
	reply, err = gpt.Completions(h.getRequestText())
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			reply = deadlineExceededText
		}
	}

	if !strings.Contains(h.msg, config.LoadConfig().ImageStartKey) {
		// 2.设置上下文，回复用户
		h.service.SetUserSessionContext(requestText, reply)
	}

	return buildUserReply(reply)
}

// getRequestText 获取请求接口的文本，要做一些清晰
func (h *UserMessageHandler) getRequestText() string {
	// 1.去除空格以及换行
	requestText := strings.TrimSpace(h.msg)
	requestText = strings.Trim(h.msg, "\n")

	// 2.获取上下文，拼接在一起，如果字符长度超出4000，截取为4000。（GPT按字符长度算），达芬奇3最大为4068，也许后续为了适应要动态进行判断。
	sessionText := h.service.GetUserSessionContext()
	if sessionText != "" {
		requestText = sessionText + "\n" + requestText
	}
	if len(requestText) >= 4000 {
		requestText = requestText[:4000]
	}

	// 3.检查用户发送文本是否包含结束标点符号
	punctuation := ",.;!?，。！？、…"
	runeRequestText := []rune(requestText)
	lastChar := string(runeRequestText[len(runeRequestText)-1:])
	if strings.Index(punctuation, lastChar) < 0 {
		requestText = requestText + "？" // 判断最后字符是否加了标点，没有的话加上句号，避免openai自动补齐引起混乱。
	}

	// 4.返回请求文本
	return requestText
}

// buildUserReply 构建用户回复
func buildUserReply(reply string) string {
	// 1.去除空格问号以及换行号，如果为空，返回一个默认值提醒用户
	textSplit := strings.Split(reply, "\n\n")
	if len(textSplit) > 1 {
		trimText := textSplit[0]
		reply = strings.Trim(reply, trimText)
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		return deadlineExceededText
	}

	reply = strings.Trim(reply, "\n")

	// 3.返回拼接好的字符串
	return reply
}
