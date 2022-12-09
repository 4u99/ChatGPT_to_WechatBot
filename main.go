package main

import (
	"ChatGPT_to_WechatBot/chatgpt"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"log"
	"strings"
)

func main() {
	startBot()
}

// startBot 登录微信
func startBot() {
	bot := openwechat.DefaultBot(openwechat.Desktop)

	// 注册消息处理函数
	bot.MessageHandler = HandlerMessage
	// 注册登陆二维码回调
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 创建热存储容器对象
	reloadStorage := openwechat.NewJsonFileHotReloadStorage("storage.json")
	// 执行热登录
	err := bot.HotLogin(reloadStorage)
	if err != nil {
		if err = bot.Login(); err != nil {
			log.Printf("login error: %v \n", err)
			return
		}
	}
	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

// HandlerMessage 处理消息
func HandlerMessage(msg *openwechat.Message) {
	if msg.IsText() { // 暂时只处理文本消息
		if msg.IsSendByGroup() {
			// 群消息
			groupMessage(msg)
		} else {
			// 私聊
			privateMessage(msg)
		}
	}
}

// groupMessage 处理群组消息
func groupMessage(msg *openwechat.Message) {
	if !msg.IsAt() {
		return
	}

	sender, err := msg.Sender()
	if err != nil {
		fmt.Println("群组消息中获取 Sender 失败:", err)
		return
	}

	groupSender, err := msg.SenderInGroup()
	if err != nil {
		log.Println("群组消息中获取 SenderInGroup 失败:", err)
		return
	}
	log.Println("群:", msg.FromUserName, "用户:", groupSender.NickName, "发送了:", msg.Content)

	// 删除@
	atText := "@" + sender.Self.NickName
	replaceMessage := strings.TrimSpace(strings.ReplaceAll(msg.Content, atText, ""))
	if replaceMessage == "" {
		return
	}

	// 获取 ChatGPT 消息
	fmt.Println("向 ChatGPT 发送:", replaceMessage)
	chatGptMessage := chatgpt.GetChatGptMessage(replaceMessage, msg.FromUserName+":"+groupSender.NickName)

	// 回复消息
	atText = "@" + groupSender.NickName
	replyText := atText + " ChatGPT回复: \n" + chatGptMessage
	_, err = msg.ReplyText(replyText)
	if err != nil {
		log.Println("发送群消息失败:", err)
	}
}

// privateMessage 处理私聊消息
func privateMessage(msg *openwechat.Message) {
	sender, err := msg.Sender()
	if err != nil {
		fmt.Println("私聊消息中获取 Sender 失败:", err)
		return
	}
	log.Println("用户:", sender.NickName, "发送了:", msg.Content)

	// 获取 ChatGPT 消息
	chatGptMessage := chatgpt.GetChatGptMessage(msg.Content, sender.ID())

	// 回复
	chatGptMessage = strings.TrimSpace(chatGptMessage)
	chatGptMessage = strings.Trim(chatGptMessage, "\n")
	chatGptMessage = "ChatGPT 回复: \n" + chatGptMessage
	_, err = msg.ReplyText(chatGptMessage)
	if err != nil {
		log.Println("发送私聊消息失败:", err)
	}
	return
}
