package cron

import (
	"fmt"
	"strings"
	"time"
	"encoding/json"
	"net/http"
	"bytes"
	"github.com/toolkits/pkg/logger"
	"github.com/weizhenqian/feishu-sender/certification"
	"github.com/weizhenqian/feishu-sender/config"
	"github.com/weizhenqian/feishu-sender/dataobj"
	"github.com/weizhenqian/feishu-sender/redisc"
)

var semaphore chan int

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
	//获取Url的值
	url := config.Get().Im.Sendurl
	//获取token
	//待优化的地方，就是这里，获取token理论可以放入内存包含过期验证。
	token := certification.GetToken()
	//初始化content
	content := genContent(message)
	//对tos做判断，解决tos为空的情况。
	if len(toslist) == 0 {
		logger.Warningf("hashid: %d: tos is empty", message.Event.HashId)
		return
	}
	//飞书的请求体，格式比较特殊，此处曾尝试struct数据类型，失败了。
	data := make(map[string]interface{})
	//飞书请求接口，具体数据可以查看url：https://open.feishu.cn/document/ukTMukTMukTM/ucDO1EjL3gTNx4yN4UTM
	data["user_ids"] = toslist
	data["msg_type"] = "text"
	data["content"] = map[string]string{"text":content}
	b, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warningf("send im fail, tos:%s, cotent:%s, error:%v", toslist, content , err)
		return
	} else {
		logger.Infof("send im succeed,tos:%s, cotent:%s", toslist, content)
	}
	defer resp.Body.Close()
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
                if message.ReadableTags == "" {
                         return fmt.Sprintf("[P%d %s] 报警信息：%s;\n监控主机：%s;\n触发时间：%s;\n报警详情：%s;\n认领报警：%s\n", message.Event.Priority, ET[message.Event.EventType], message.Event.Sname, message.ReadableEndpoint, parseEtime(message.Event.Etime), message.EventLink, message.ClaimLink)
                } else {
                        return fmt.Sprintf("[P%d %s] 报警信息：%s;\n监控主机：%s;\n简略信息：%s;\n触发时间：%s;\n报警详情：%s;\n认领报警：%s\n", message.Event.Priority, ET[message.Event.EventType], message.Event.Sname, message.ReadableEndpoint, message.ReadableTags,parseEtime(message.Event.Etime), message.EventLink, message.ClaimLink)
                }
        } else {
                if message.ReadableTags == "" {
                        return fmt.Sprintf("[P%d %s] 报警信息：%s;\n监控主机：%s;\n触发时间：%s;\n", message.Event.Priority, ET[message.Event.EventType], message.Event.Sname, message.ReadableEndpoint, parseEtime(message.Event.Etime))
                } else {
                        return fmt.Sprintf("[P%d %s] 报警信息：%s;\n监控主机：%s;\n简略信息：%s;\n触发时间：%s;\n", message.Event.Priority, ET[message.Event.EventType], message.Event.Sname, message.ReadableEndpoint, message.ReadableTags, parseEtime(message.Event.Etime))
                }
	}
}

func parseEtime(etime int64) string {
	t := time.Unix(etime, 0)
	return t.Format("2006-01-02 15:04:05")
}
