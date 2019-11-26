package api

import (
	"encoding/json"

	"github.com/bingoohuang/goreflect"

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

func (t *NotifyConfigCache) write(k string, v interface{}, writeSnapshot bool) error {
	t.C.SetDefault(k, v)

	var bytes []byte

	var err error

	if bytes, err = json.Marshal(v); err != nil {
		return err
	}

	if writeSnapshot {
		return t.Snapshot.Write(k+".json", bytes)
	}

	return nil
}

func (t *NotifyConfigCache) Write(k string, v *NotifyConfig, writeSnapshot bool) error {
	return t.write(k, v, writeSnapshot)
}

const redlistKey = "_redlist_"

func (t *NotifyConfigCache) WriteRedList(v RedList, writeSnapshot bool) error {
	return t.write(redlistKey, v, writeSnapshot)
}

func (t *NotifyConfigCache) ReadRedList() (redlist RedList) {
	key := redlistKey
	v, found := t.C.Get(key)

	if found {
		return v.(RedList)
	}

	content, _ := t.Snapshot.Read(key + ".json")
	if len(content) == 0 {
		return redlist
	}

	if err := json.Unmarshal(content, &redlist); err != nil {
		logrus.Warnf("json Unmarshal failed %v", err)
		return redlist
	}

	if err := t.WriteRedList(redlist, false); err != nil {
		logrus.Warnf("write snapshot failed %v", err)
	}

	return redlist
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
	keys := goreflect.MapKeys(items)

	for _, ki := range keys {
		vi := items[ki]
		if config, ok := vi.Object.(*NotifyConfig); ok {
			fn(ki, config)
		}
	}
}

func (t *NotifyConfigCache) CleanAll() {
	t.C.Flush()
}
