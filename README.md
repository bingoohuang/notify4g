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

API document style refers to [White House Web API Standards](https://github.com/WhiteHouse/api-standards).

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
<details><summary>GET /raw/qcloudsms</summary>
<p>
Response body:

```json
{
    "config": {
        "sdkappid": "VIsOIVUTXKvmznGCfpklQBsHl",
        "appkey": "jyuRWrnndYwTzEQIDtpaulCEv",
        "tplID": 58,
        "sign": "",
        "tmplVarNames": [
            "NtAOrrDyTQZprXHlRyMKIQVrJ"
        ]
    },
    "data": {
        "params": [
            "DfQeutEzaCfShlItCeaEkTUGF",
            "DZiVFEPPlMANVxGwaCVjypmXA"
        ],
        "mobiles": [
            "15923459113",
            "18923435937"
        ]
    }
}
```

</p>
</details>
<details><summary>GET /raw/qcloudvoice</summary>
<p>
Response body:

```json
{
    "config": {
        "sdkappid": "dqpaGfzwZsdYPeOyCsiCnHuLe",
        "appkey": "HQFbAYSZWVMAhuzBkneOovYpv",
        "tplID": 39,
        "playTimes": 49,
        "tmplVarNames": [
            "KyzjUzrBFcqQjedfJRHYoDbOG",
            "nflOHIkugcnZOrqBkSazNWfPP"
        ]
    },
    "data": {
        "params": {
            "RIsCQnfJqlpSCwrkkFbdBFIFj": "XXcLBedVQEUDCnYApsnqfVPTL"
        },
        "mobile": "13534814833"
    }
}
```

</p>
</details>
<details><summary>GET /raw/qywx</summary>
<p>
Response body:

```json
{
    "config": {
        "corpID": "uCgrmJMtqLPBCFhsvjTArsMmL",
        "corpSecret": "gRHZGuimGqaWdWaBWJwkTAShU",
        "agentID": "GjqWAhwRbpeHnQNxTNgmJjnxD"
    },
    "data": {
        "msg": "SlooidCOblAgkzyWhxDcYtLJJ",
        "userIds": [
            "yFeNmhPfjtisROYMvzGXHlQpd",
            "CUSPjJkWEEfDDKDOfOhAXkqgJ"
        ]
    }
}
```

</p>
</details>
<details><summary>GET /raw/mail</summary>
<p>
Response body:

```json
{
    "config": {
        "smtpAddr": "xaQHabaoaboiqLQkrhnMSwTGo",
        "smtpPort": 94,
        "from": "CEVyoTJ@zTADH.biz",
        "username": "gQVadOOpmwpHnlIyfsCCBulVP",
        "pass": "NvMQhtbtbJgCkOErmOqWRCSKa"
    },
    "data": {
        "subject": "kfYENjIqRgtAsNATTewtSQJtK",
        "message": "nbUGDoWZCCUeCgZnqaHOhlDUc",
        "to": [
            "RyBZMmL@NbhCr.net",
            "fKggIDs@WHkmM.net"
        ]
    }
}
```

</p>
</details>
<details><summary>GET /raw/sms</summary>
<p>
Response body:

```json
{
    "config": {
        "configIds": [
            "NCCNkSbvLCcEBYPRpErzuHOzu",
            "qJoxwNVZfRxOytgzHbfYLSnNg"
        ],
        "random": false,
        "retry": 0
    },
    "data": {
        "templateParams": {
            "CAatGFIenVaglyBHaqLGDVNDm": "HmVmUYwDzaKJZvYwyMqYAAowJ",
            "uCbSRZgyNkgIntizzDrIHVOiy": "TNvBcIHcgrHoiKEGjEsktAKmn"
        },
        "mobiles": [
            "14509804092"
        ],
        "retry": 0
    }
}
```

</p>
</details>