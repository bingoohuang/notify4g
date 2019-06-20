package api

import (
	"encoding/json"

	"github.com/bingoohuang/gou"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type NotifyConfigCache struct {
	C        *cache.Cache
	Snapshot SnapshotService
}

func NewCache(snapshotDir string) *NotifyConfigCache {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	c := &NotifyConfigCache{C: cache.New(cache.NoExpiration, cache.NoExpiration)}

	if err := c.Snapshot.Init(snapshotDir); err != nil {
		logrus.Panic(err)
	}

	if err := c.Snapshot.RecoverCache(c); err != nil {
		logrus.Panic(err)
	}

	return c
}

func (t *NotifyConfigCache) Write(k string, v *NotifyConfig, writeCache bool) error {
	t.C.SetDefault(k, v)

	var bytes []byte
	var err error
	if bytes, err = json.Marshal(v); err != nil {
		return err
	}

	if writeCache {
		return t.Snapshot.Write(k+".json", bytes)
	}

	return nil
}

func (t *NotifyConfigCache) Read(key string) *NotifyConfig {
	v, found := t.C.Get(key)
	if found {
		return v.(*NotifyConfig)
	}

	content, _ := t.Snapshot.Read(key + ".json")
	if len(content) != 0 {
		c, _ := ParseNotifyConfig(content)
		if c != nil {
			if err := t.Write(key, c, false); err != nil {
				logrus.Warnf("write snapshot failed %v", err)
			}
			return c
		}
	}

	return nil
}

func (t *NotifyConfigCache) Delete(key string) {
	t.C.Delete(key)
	if err := t.Snapshot.Delete(key + ".json"); err != nil {
		logrus.Warnf("delete snapshot failed %v", err)
	}

}

func (t *NotifyConfigCache) Walk(fn func(k string, v *NotifyConfig)) {
	items := t.C.Items()
	keys := gou.MapKeys(items).([]string)

	for _, ki := range keys {
		vi := items[ki]
		fn(ki, vi.Object.(*NotifyConfig))
	}
}

func (t *NotifyConfigCache) CleanAll() {
	t.C.Flush()
}
