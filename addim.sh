#!/bin/bash
#Usage:根据夜莺API和飞书API自动获取用户ID并回写到夜莺的用户信息。
#注意：注册飞书的用户手机号需要和夜莺填写的手机号一致。
#首先通过夜莺获取用户的基础信息，重点为：手机号
#根据飞书API中基于手机号查询用户信息获取用户的飞书ID。
#根据夜莺监控API重写到用户信息里。
#由于；逻辑比较简单，就没有太注重格式和逻辑，就写了一些简单的调用，可以优化。

###执行前，需要检测飞书API是否有根据用户手机查找用户ID的权限！很重要！###

#定义说明
Usage='执行前，需要检测夜莺API是否有超管用户，飞书API是否有根据用户手机查找用户ID的权限！很重要！'

#定义变量
#夜莺API
N9EAPI=''  #格式为: 用户名:密码 例如：user:password
N9ESERVERIP=''  #服务地址
N9ESERVERPORT=''  #服务端口号
USERNUMS='100'  #显示用户数
#飞书APP
app_id=''
app_secrt=''

#定义函数
#检测运行环境
function check_env(){
  #很多命令都是通过jq调用的，如果没有jq环境则无法运行
  JqFlag=$(command -V jq >/dev/null 2>&1; echo -n "$?")
  if [ "${JqFlag}" -ne 0 ];then
    echo "jq命令不存在，请自行安装。"
  fi
}

#获取用户信息，以特定格式存储到变量中。
function get_userinfos(){
  userinfos=$(curl -s -u "${N9EAPI}" "http://${N9ESERVERIP}:${N9ESERVERPORT}/api/portal/user?limit=100&p=1" | jq -r ".dat.list[] | [.id,.dispname,.phone,.email,.is_root]"|tr -s '\n' ' ' |sed -e 's/]/],/g' -e 's/, $//g')
  #获取用户信息的长度（即用户数）
  userlen=$(echo -n "[${userinfos}]" |jq .|jq 'length')
}

#获取token
function get_token(){
  #获取token
  token=$(curl -s -XPOST -H "Content-Type:application/json" -d "{\"app_id\":\"${app_id}\",\"app_secret\":\"${app_secrt}\"}" 'https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal/' |jq -r .app_access_token)
}

#获取单个用户的ID
function get_userid(){
  phone=$1
  uid=$(curl -s -XGET -H "Content-Type:application/json" -H "Authorization:Bearer ${token}" "https://open.feishu.cn/open-apis/user/v1/batch_get_id?mobiles=${phone}"|jq -r ".data.mobile_users.\"${phone}\"[0].user_id")
}

#更改夜莺用户信息
function change_n9eim(){
  id=$1
  name=$2
  phone=$3
  email=$4
  im=$5
  isroot=$6
  curl -XPUT -H "Content-Type:application/json" -u "${N9EAPI}" -d "{\"dispname\":\"${name}\",\"phone\":\"${phone}\",\"email\":\"${email}\",\"im\":\"${im}\",\"is_root\":${isroot}}" "http://${N9ESERVERIP}:${N9ESERVERPORT}/api/portal//user/${id}/profile" 
}


#循环执行修改用户信息
function do_body(){
  for((i=0;i<${userlen};i++))
  do
    n9eid=$(echo -n "[${userinfos}]"|jq -r .[$i][0])
    n9ename=$(echo -n "[${userinfos}]"|jq -r .[$i][1])
    n9ephone=$(echo -n "[${userinfos}]"|jq -r .[$i][2])
    n9email=$(echo -n "[${userinfos}]"|jq -r .[$i][3])
    n9eisroot=$(echo -n "[${userinfos}]"|jq -r .[$i][4])
    get_userid $n9ephone
    change_n9eim $n9eid $n9ename $n9ephone $n9email $uid $n9eisroot 
  done
}

function main(){
  if [ $# -ne 1 ];then
    echo "请执行:$0 check"
    check_env
    exit
  fi
  if [ "$1"x != 'start'x ];then
    echo "$Usage" 
    echo "如果检测完毕，请执行$0 start"
    exit
  fi
  get_userinfos
  get_token
  do_body
}
main $@
