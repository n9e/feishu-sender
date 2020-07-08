package cron

import (
	"fmt"
	"strings"
	"time"

	"github.com/toolkits/net/httplib"
	"github.com/toolkits/pkg/logger"
	"github.com/weizhenqian/im-sender/config"
	"github.com/weizhenqian/im-sender/dataobj"
	"github.com/weizhenqian/im-sender/redisc"
)

var semaphore chan int

func SendVoices() {
	c := config.Get()

	//定义发送限制
	semaphore = make(chan int, c.Consumer.Worker)
	//循环读取redis的值
	for {
		messages := redisc.Pop(1, c.Consumer.Queue)
		if len(messages) == 0 {
			time.Sleep(time.Duration(300) * time.Millisecond)
			continue
		}
		//读取到值，则调用sendSmss发送
		sendVoices(messages)
	}
}

func sendVoices(messages []*dataobj.Message) {
	//读取messages的值，获取单个信息，调用sendSms发送
	for _, message := range messages {
		semaphore <- 1
		go sendVoice(message)
	}
}

func sendVoice(message *dataobj.Message) {
	defer func() {
		<-semaphore
	}()

	//初始化tos的值（避免存在空值的情况）
	cnt := len(message.Tos)
	toslist := make([]string, 0, cnt)
	for i := 0; i < cnt; i++ {
		item := strings.TrimSpace(message.Tos[i])
		if item == "" {
			continue
		}
		toslist = append(toslist, item)
	}
	tos := strings.Replace(strings.Trim(fmt.Sprint(toslist), "[]"), " ", ",", -1)
	//获取Url的值
	url := config.Get().Sms.Url

	//初始化content
	content := genContent(message)
	data := []string{"text": content}
	if len(tos) == 0 {
		logger.Warningf("hashid: %d: tos is empty", message.Event.HashId)
		return
	}

	r := httplib.Post(url).SetTimeout(5*time.Second, 30*time.Second)
	req.Header("Content-Type", "application/json")
	req.Header("Authorization", "Bearer t-a2b4116b22cc833c06c6476dd76d822133965ab2")
	r.Param("user_ids", tos)
	r.Param("msg_type", "text")
	r.Param("content", data)
	_, err := r.String()
	if err != nil {
		logger.Warningf("send sms fail, tos:%s, cotent:%s, error:%v", tos, content, err)
		return
	} else {
		logger.Infof("send sms succeed,tos:%s, cotent:%s", tos, content)
	}
}

var ET = map[string]string{
	"alert":    "告警",
	"recovery": "恢复",
}

func genContent(message *dataobj.Message) string {
	content := ""
	if message.IsUpgrade {
		content = "[报警已升级]" + content
	}
	//此处新增if判断，解决认领报警为空的情况（恢复报警不包含认领报警连接）
	if message.Event.EventType == "alert" {
		return fmt.Sprintf("[P%d %s]报警信息:%s; 主机:%s; 触发时间：%s; 报警详情:%s; 认领报警:%s", message.Event.Priority, ET[message.Event.EventType], message.Event.Sname, message.ReadableEndpoint, parseEtime(message.Event.Etime), message.EventLink, message.ClaimLink)
	} else {
		return fmt.Sprintf("[P%d %s]报警信息:%s; 主机:%s; 触发时间：%s; 报警详情:%s;", message.Event.Priority, ET[message.Event.EventType], message.Event.Sname, message.ReadableEndpoint, parseEtime(message.Event.Etime), message.EventLink)
	}
}

func parseEtime(etime int64) string {
	t := time.Unix(etime, 0)
	return t.Format("2006-01-02 15:04:05")
}
