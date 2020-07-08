package certification

import (
	"fmt"
	"github.com/toolkits/net/httplib"
	"github.com/weizhenqian/im-sender/config"
)

func GetToken() string {
	addid := config.Get().App.Appid	
	appsecret := config.Get().App.Appsecret
	url := 	config.Get().Im.Tokenurl
	req := httplib.Post(url)
	req.Header("Content-Type","application/json")
	req.Param("app_id", appid)
	req.Param("app_secret", appsecret)
	var str interface{}
        err := req.ToJson(&str)
        if err != nil {
                fmt.Println(err)
        }
    token := str.(map[string]interface{})["app_access_token"]
	return token
}