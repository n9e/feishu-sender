package cron

import (
	"fmt"
	"strings"
	"time"
	"encoding/json"

	"github.com/toolkits/net/httplib"
	"github.com/toolkits/pkg/logger"
	"github.com/weizhenqian/im-sender/certification"
	"github.com/weizhenqian/im-sender/config"
	"github.com/weizhenqian/im-sender/dataobj"
	"github.com/weizhenqian/im-sender/redisc"
)

var semaphore chan int

type data struct {
	x string
	y string
}

func SendIms() {
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
		//读取到值，则调用sendIms发送
		sendIms(messages)
	}
}

func sendIms(messages []*dataobj.Message) {
	//读取messages的值，获取单个信息，调用sendIm发送
	for _, message := range messages {
		semaphore <- 1
		go sendIm(message)
	}
}

func sendIm(message *dataobj.Message) {
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
	tostemp,e := json.Marshal(toslist)
	if e != nil {
		panic(e)
	}
	tos := fmt.Sprintf("%s", tostemp)
	//获取Url的值
	url := config.Get().Im.Sendurl
	//获取token
	token := certification.GetToken()
	//初始化content
	content := genContent(message)
	datatemp := data{"text",content}
	body := fmt.Sprintf("{%s:%s}", datatemp.x,datatemp.y) 
	if len(tos) == 0 {
		logger.Warningf("hashid: %d: tos is empty", message.Event.HashId)
		return
	}

	r := httplib.Post(url).SetTimeout(5*time.Second, 30*time.Second)
	r.Header("Content-Type", "application/json")
	r.Header("Authorization", token)
	r.Param("user_ids", tos)
	r.Param("msg_type", "text")
	r.Param("content", body)
	str, err := r.String()
	if err != nil {
		logger.Warningf("send im fail, tos:%s, cotent:%s, error:%v", tos, body, err)
		logger.Warningf("str:s%",str)
		return
	} else {
		logger.Infof("send im succeed,tos:%s, cotent:%s", tos, body)
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
