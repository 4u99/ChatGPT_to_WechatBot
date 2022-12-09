package chatgpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"
)

func GetChatGptMessage(requestText string, openId string) string {
	fmt.Println("向 ChatGPT 发送:", requestText)
	chatGptMessage := DefaultGPT.SendMsg(requestText, openId)
	chatGptMessage = strings.TrimSpace(chatGptMessage)
	chatGptMessage = strings.Trim(chatGptMessage, "\n")
	return chatGptMessage
}

var (
	DefaultGPT  = newChatGPT()
	userInfoMap = make(map[string]*userInfo)
	lock        = sync.Mutex{}
)

type ChatGPT struct {
	ok            bool
	authorization string
	sessionToken  string
}

type userInfo struct {
	parentID       string
	conversationId interface{}
	ttl            time.Time
}

func newChatGPT() *ChatGPT {
	sessionToken, err := os.ReadFile("sessionToken")
	if err != nil {
		log.Println("读取 sessionToken 文件失败:", err)
		exit()
	}
	if len(sessionToken) < 100 {
		log.Println("你应该忘了配置 sessionToken")
		exit()
	}
	gpt := &ChatGPT{
		sessionToken: *(*string)(unsafe.Pointer(&sessionToken)),
	}
	if !gpt.updateSessionToken() {
		exit()
	}

	// 每 10 分钟更新一次 sessionToken
	go func() {
		for range time.Tick(10 * time.Minute) {
			gpt.updateSessionToken()
		}
	}()
	return gpt
}

func (c *ChatGPT) updateSessionToken() bool {
	c.ok = false
	session, err := http.NewRequest("GET", "https://chat.openai.com/api/auth/session", nil)
	if err != nil {
		log.Println("更新 Token 调用 NewRequest 失败:", err)
		return false
	}
	session.AddCookie(&http.Cookie{
		Name:  "__Secure-next-auth.session-token",
		Value: c.sessionToken,
	})
	session.AddCookie(&http.Cookie{
		Name:  "__Secure-next-auth.callback-url",
		Value: "https://chat.openai.com/",
	})
	session.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")
	resp, err := http.DefaultClient.Do(session)
	if err != nil {
		log.Println("更新 Token 调用接口失败:", err)
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "__Secure-next-auth.session-token" {
			c.sessionToken = cookie.Value
			_ = os.WriteFile("sessionToken", []byte(cookie.Value), 0600)
			break
		}
	}
	var accessToken map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&accessToken)
	if err != nil {
		log.Println("更新 Token 解析响应数据失败:", err)
		return false
	}
	c.authorization = accessToken["accessToken"].(string)
	log.Println("sessionToken 更新成功")
	c.ok = true
	return true
}

func (c *ChatGPT) SendMsg(msg, openId string) string {
	if !c.ok {
		return "ChatGPT 没有初始化成功"
	}

	lock.Lock()
	defer lock.Unlock()

	info, ok := userInfoMap[openId]
	if !ok || info.ttl.Before(time.Now()) {
		log.Printf("用户 %s 启动新的对话\n", openId)
		info = &userInfo{
			parentID:       uuid.New().String(),
			conversationId: nil,
		}
		userInfoMap[openId] = info
	} else {
		log.Printf("用户 %s 继续对话\n", openId)
	}
	info.ttl = time.Now().Add(5 * time.Minute)

	// 发送请求
	req, err := http.NewRequest("POST", "https://chat.openai.com/backend-api/conversation", CreateChatReqBody(msg, info.parentID, info.conversationId))
	if err != nil {
		log.Println("调用 ChatGPT 的 NewRequestWithContext 异常:", err)
		return "服务器异常, 请稍后再试"
	}
	req.Header.Set("Host", "chat.openai.com")
	req.Header.Set("Authorization", "Bearer "+c.authorization)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Openai-Assistant-App-Id", "")
	req.Header.Set("Connection", "close")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://chat.openai.com/chat")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("调用 ChatGPT 接口异常:", err)
		return "服务器异常, 请稍后再试"
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取 ChatGPT 数据异常", err)
		return "服务器异常, 请稍后再试"
	}
	line := bytes.Split(bodyBytes, []byte("\n\n"))
	if len(line) < 2 {
		log.Println("回复数据格式不对:", string(bodyBytes))
		return "服务器异常, 请稍后再试"
	}
	endBlock := line[len(line)-3][6:]
	res := ToChatRes(endBlock)
	info.conversationId = res.ConversationId
	info.parentID = res.Message.Id
	if len(res.Message.Content.Parts) > 0 {
		return res.Message.Content.Parts[0]
	} else {
		return "没有获取到..."
	}
}

func exit() {
	log.Println("请输入任意字符退出程序")
	_, _ = os.Stdin.Read([]byte{0})
	os.Exit(0)
}
