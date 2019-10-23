package api

import (
	"fmt"

	"github.com/bingoohuang/gou"
)

type QcloudBase struct {
	Sdkappid string `json:"sdkappid"`
	Appkey   string `json:"appkey"`
}

func (q QcloudBase) CreateSignature(rand string, t int64, nums ...string) string {
	src := fmt.Sprintf("appkey=%s&random=%s&time=%d&mobile=%s",
		q.Appkey, rand, t, gou.JoinNonEmpty(",", nums...))
	s, _ := gou.Sha256(src)

	return s
}
