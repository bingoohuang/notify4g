package api

import (
	"github.com/sirupsen/logrus"
	"strings"
)

// RedList 表示红名单，红名单里面配置的人，不可以被“骚扰”
type RedList struct {
	Mobiles     []string `json:"mobiles"`
	Mails       []string `json:"mails"`
	QywxUserIds []string `json:"qywxUserIds"`
}

// redList 内部实现，为了方便使用
type redList struct {
	mobiles     map[string]bool
	mails       map[string]bool
	qywxUserIds map[string]bool
}

func TrimSpaceToLower(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// SliceToMap 切片转换为 map[string]bool
func SliceToMap(slice []string) map[string]bool {
	m := make(map[string]bool)

	for _, s := range slice {
		if ss := TrimSpaceToLower(s); ss != "" {
			m[ss] = true
		}
	}

	return m
}

// FilterSlice 根据m过滤切片slice，返回剩余项目
func FilterSlice(slice []string, m map[string]bool) []string {
	ret := make([]string, 0, len(slice))
	filtered := make([]string, 0)

	for _, k := range slice {
		if k = strings.TrimSpace(k); k == "" {
			continue
		}

		if _, ok := m[strings.ToLower(k)]; !ok {
			ret = append(ret, k)
		} else {
			filtered = append(filtered, k)
		}
	}

	if len(filtered) > 0 {
		logrus.Warnf("redlist filtered %v", filtered)
	}

	return ret
}

func (l RedList) prepare() redList {
	var r redList

	r.mobiles = SliceToMap(l.Mobiles)
	r.mails = SliceToMap(l.Mails)
	r.qywxUserIds = SliceToMap(l.QywxUserIds)

	return r
}

func (l redList) FilterMobiles(mobiles []string) []string { return FilterSlice(mobiles, l.mobiles) }
func (l redList) FilterMails(mails []string) []string     { return FilterSlice(mails, l.mails) }
func (l redList) FilterQywxUserIds(ids []string) []string { return FilterSlice(ids, l.qywxUserIds) }
