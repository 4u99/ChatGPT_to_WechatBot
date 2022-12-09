# ChatGPT_to_WechatBot
基于ChatGPT的微信机器人

# 引用
基于[openwechat](https://github.com/eatmoreapple/openwechat)的微信机器人

ChatGPT的接口参考了[wechat-chatGPT](https://github.com/gtoxlili/wechat-chatGPT)

# 使用方法
## sessionToken
打开浏览器并进入 ChatGPT 页面, 在请求中复制 cookie 中 __Secure-next-auth.session-token 的值 到 sessionToken 文件

![GetSessionToken](https://github.com/lihongbin99/ChatGPT_to_WechatBot/blob/master/static/token.png?raw=true)

# 编译
```
go build -o WechatBot.exe main.go
```