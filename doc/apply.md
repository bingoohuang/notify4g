通知配置

# 企业微信

参数 |是否必须 |	说明
----|--------|-----
corpid|是|企业ID，获取方式参考：术语说明-corpid
corpsecret|是|企业应用的凭证密钥，获取方式参考：术语说明-secret
agentid |	是	 |企业应用的id，整型。可在应用的设置页面查看

1. [corpid、corpsecret说明](https://work.weixin.qq.com/api/doc#90000/90135/91039)
1. [agentid说明](https://work.weixin.qq.com/api/doc#90000/90135/90250/%E6%96%87%E6%9C%AC%E6%B6%88%E6%81%AF)
1. [企业自建应用](https://open.work.weixin.qq.com/wwopen/helpguide/detail?t=selfBuildApp)

```json
{
    "config": {
        "corpID": "wwe4f0f9b348e7779",
        "corpSecret": "Vt7qZmgTkIgx1z1vsLsmPGzqbzEOd5sHQhqsWk",
        "agentID": "1000002"
    }
}
```

![image](https://user-images.githubusercontent.com/1940588/60090421-1ae9ab80-9775-11e9-8e1a-6805fb7daff5.png)

![image](https://user-images.githubusercontent.com/1940588/60089776-c134b180-9773-11e9-9e42-827ffde4af6c.png)

![image](https://user-images.githubusercontent.com/1940588/60089839-e7f2e800-9773-11e9-9b0a-4906119a2305.png)

# 钉钉自定义机器人


参数 |是否必须 |	说明
----|--------|-----
accessToken|是| 自定义机器人hook地址中的accessToken


示例:

```json
{
    "config": {
      "accessToken": "e9aec9bf0429505e7c16ba7090b860694b4835c55c69c25113bed7ab46da5"
    }
}
```

1. [钉钉自定义机器人](https://open-doc.dingtalk.com/microapp/serverapi2/qf2nxq)

![image](https://user-images.githubusercontent.com/1940588/60089297-c1807d00-9772-11e9-9ad5-11f31b9a634f.png)



# 阿里云短信

参数 |是否必须 |	说明
----|--------|-----
accessKeyID|是| AccessKeyId用于标识用户。
acessKeySecret|是| AccessKeySecret是用来验证用户的密钥。AccessKeySecret必须保密。
TemplateCode|是| 短信模板CODE。


1. [如何获取AccessKey ID和AccessKey Secret](https://help.aliyun.com/knowledge_detail/48699.html)
1. [短信服务 > API参考 > 发送短信 > SendBatchSms](https://help.aliyun.com/document_detail/102364.html?spm=a2c4g.11186623.6.615.451856e0wbes4c)
1. [申请短信模版帮助](https://help.aliyun.com/document_detail/55330.html)

示例:

```json
{
    "config": {
        "accessKeyID": "LTAIaB1FFgE3na",
        "acessKeySecret": "K4HaHplXvxA1atHfKp0O0M44NdD",
        "templateCode": "SMS_138125087",
        "signName": "告警"
    }
}
```

模板示例: `应用:${appName} 监控埋点:${warnSrc} 在近${withMinutes}分钟内发生${warning}, 其中最高${max}, 最低${min}`


![image](https://user-images.githubusercontent.com/1940588/60090827-d3afea80-9775-11e9-99a5-a42e6d9420b4.png)

# 腾讯云短信

参数 |是否必须 |	说明
----|--------|-----
sdkappid|是| AccessKeyId用于标识用户。
appkey|是| AccessKeySecret是用来验证用户的密钥。AccessKeySecret必须保密。
tplID|是| 模板ID，在 控制台 审核通过的模板 ID。

示例:

```json
{
    "config": {
        "sdkappid": "14000840",
        "appkey": "71055d15873371cd13d7b15c89341",
        "tplID": 157749
    }
}
```

模板示例: `应用:{1} 监控埋点:{2} 在近{3}分钟内发生{4}, 其中最高{5}, 最低{6}`

1. [手把手教你使用腾讯云短信服务——开发者视角](https://cloud.tencent.com/developer/article/1154647)
1. [指定模板群发短信](https://cloud.tencent.com/document/product/382/5977)


![image](https://user-images.githubusercontent.com/1940588/60091843-14a8fe80-9778-11e9-953c-039fc15263ce.png)


# 邮箱

参数 |是否必须 |	说明
----|--------|-----
smtpAddr|是| 邮箱的SMTP地址。
smtpPort|是| 邮箱的SMTP端口。
from|是| 发出人。
username|是|邮箱登录用户名
pass|是|邮箱登录密码

示例:

```json
{
    "config": {
        "smtpAddr": "mail.amail.cn",
        "smtpPort": 25,
        "from": "i@bj.cn",
        "username": "is@bj.cn",
        "pass": "xba"
    }
}
```

