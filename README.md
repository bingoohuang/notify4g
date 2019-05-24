# notify4g
notify api for sms/voice/qywx/mail/dingtalk

## build

1. `go get github.com/bingoohuang/statiq`
1. `./buildres.sh`
1. `statiq -src=res`
1. `go fmt ./...; go build`

## snapshots

![image](doc/snapshot20190523141355.png)


## Request & Response Examples

### API Resources

* [GET /raw/:channel](#get-rawchannel)
* POST /raw/:channel
* GET /config/:configID
* GET /config/:configID/:channel
* POST /config/:configID
* DELETE /config/:configID
* GET /notify/:configID
* POST /notify/:configID

### GET /raw/:channel

<details><summary>GET /raw/aliyunsms</summary>
<p>
Response body:

```json
{
    "config": {
        "accessKeyID": "BvXitpxZTiQPBPJNHzKEyUtZX",
        "acessKeySecret": "EaouuhLQvkjvBpcqjySEaDtZp",
        "templateCode": "jHsEdCAyjQiwwKfKTIAyMwhLd",
        "signName": ""
    },
    "data": {
        "templateCode": "",
        "templateParams": {
            "VIbvxqDAKzYvRCOugkfSTdBii": "HUzqgIrkvrpoUvwfOnPkYWJCc",
            "nSodfKtCBsGOdFcFIfhRfKkQD": "MhBpCGIHwwFSTrekZojpWHHRj"
        },
        "signName": "",
        "mobiles": [
            "13640030119"
        ]
    }
}
```

</p>
</details>
<details><summary>GET /raw/dingtalkrobot</summary>
<p>
Response body:

```json
{
    "config": {
        "accessToken": "bmluXMmkzbKJXhHvRYnPWEFon"
    },
    "data": {
        "message": "uxuNAGIvNfwPHCppEJGAFbbJb",
        "atMobiles": [
            "16231720931",
            "12123690368"
        ],
        "atAll": true
    }
}
```

</p>
</details>