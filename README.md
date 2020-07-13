# feishu-sender

Nightingale的理念，是将告警事件扔到redis里就不管了，接下来由各种sender来读取redis里的事件并发送，毕竟发送报警的方式太多了，适配起来比较费劲，希望社区同仁能够共建。

由于我公司内部聊天工具为飞书，查看社区并未包含飞书的相关报警发送方式，特在此分享自己构建飞书机器人及编写sender模块遇到的一些坑。

## create a rebot

发送是基于机器人的方式发送，肯定需要先构建一个机器人。
具体可以查看飞书介绍文档，在此就不过多赘述了。
飞书构建机器人：
https://open.feishu.cn/document/ukTMukTMukTM/uATM04CMxQjLwEDN
注：机器人的发送消息API对golang真的特别不友好，请求体的问题困扰了我一整天。

## get user id

获取用户的id需要调用接口，具体逻辑在addim.sh脚本里面写得很清楚了。可以```cat addim.sh```看一下。
如果不想看脚本，在脚本中配置上API的变量，执行即可```sh addim.sh check```。
```bash
#夜莺API
N9EAPI=''  #格式为: 用户名:密码 例如：user:password
N9ESERVERIP=''  #服务地址
N9ESERVERPORT=''  #服务端口号
USERNUMS='100'  #显示用户数
#飞书APP
app_id=''
app_secrt=''
```

## compile

```bash
cd $GOPATH/src
mkdir -p github.com/n9e
cd github.com/n9e
git clone https://github.com/n9e/feishu-sender.git
cd feishu-sender
./control build
```

如上编译完就可以拿到二进制了。

## configuration

读取告警事件，自然要给出redis的连接地址；发送邮件，自然要给出smtp配置；直接修改etc/mail-sender.yml即可

## pack

编译完成之后可以打个包扔到线上去跑，将二进制和配置文件打包即可：

```bash
tar zcvf feishu-sender.tar.gz feishu-sender etc/feishu-sender.yml
```

## test

木有写测试

## run

如果测试邮件发送没问题，扔到线上跑吧，使用systemd或者supervisor之类的托管起来，systemd的配置实例：


```
$ cat feishu-sender.service
[Unit]
Description=Nightingale feishu sender
After=network-online.target
Wants=network-online.target

[Service]
User=root
Group=root

Type=simple
ExecStart=/home/n9e/feishu-sender
WorkingDirectory=/home/n9e

Restart=always
RestartSec=1
StartLimitInterval=0

[Install]
WantedBy=multi-user.target
```