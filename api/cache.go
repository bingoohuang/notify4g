package api

import (
	"encoding/json"
	"github.com/bingoohuang/gokv/pkg/sqlc"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type NotifyConfigCache struct {
	C        *cache.Cache
	Snapshot SnapshotService
	client   *sqlc.Client
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

	dsn := viper.GetString("DSN")
	if dsn == "" {
		return c
	}

	interval, _ := time.ParseDuration(viper.GetString("RefreshInterval"))

	c.client = sqlc.NewClient(sqlc.Config{
		DriverName:      viper.GetString("DriverName"),
		DataSourceName:  dsn,
		RefreshInterval: interval,
		GetSQL:          viper.GetString("GetSQL"),
		SetSQL:          viper.GetString("SetSQL"),
		DelSQL:          viper.GetString("DelSQL"),
	})

	if _, err := c.client.All(); err != nil {
		logrus.Warnf("failed read all %v", err)
	}

	return c
}

func (t *NotifyConfigCache) write(k string, v interface{}, writeSnapshot bool) error {
	t.C.SetDefault(k, v)

	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if writeSnapshot {
		if t.client != nil {
			if err := t.client.Set(k, string(bytes)); err != nil {
				logrus.Warnf("failed to set %s=%s error %v", k, bytes, err)
			}
		}

		return t.Snapshot.Write(k+".json", bytes)
	}

	return nil
}

func (t *NotifyConfigCache) Write(k string, v *NotifyConfig, persist bool) error {
	return t.write(k, v, persist)
}

const redlistKey = "_redlist_"

func (t *NotifyConfigCache) WriteRedList(v RedList, persist bool) error {
	return t.write(redlistKey, v, persist)
}

func (t *NotifyConfigCache) ReadRedList() (redlist RedList) {
	key := redlistKey
	v, found := t.C.Get(key)

	if found {
		return v.(RedList)
	}

	var content []byte

	if v, ok := t.readKeyByClient(key); ok {
		content = []byte(v)
	} else {
		content, _ = t.Snapshot.Read(key + ".json")
	}

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

func (t *NotifyConfigCache) readKeyByClient(key string) (string, bool) {
	if t.client == nil {
		return "", false
	}

	found, v, err := t.client.Get(key)
	if err != nil {
		log.Printf("E! fail to get by key %s: %v", key, err)
		return "", false
	}

	return v, found
}

func (t *NotifyConfigCache) readByClient(key string) *NotifyConfig {
	v, found := t.readKeyByClient(key)
	if !found {
		return nil
	}

	nc, err := ParseNotifyConfig([]byte(v))
	if err != nil {
		log.Printf("E! fail to get by key %s: %v", key, err)

		return nil
	}

	return nc
}

func (t *NotifyConfigCache) Read(key string) *NotifyConfig {
	if v := t.readByClient(key); v != nil {
		return v
	}

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
func (t *NotifyConfigCache) delByClient(key string) {
	if t.client == nil {
		return
	}

	if err := t.client.Del(key); err != nil {
		logrus.Warnf("failed to read key %s err %v", key, err)
	}
}

func (t *NotifyConfigCache) Delete(key string) {
	t.delByClient(key)
	t.C.Delete(key)

	if err := t.Snapshot.Delete(key + ".json"); err != nil {
		logrus.Warnf("delete snapshot failed %v", err)
	}
}

func (t *NotifyConfigCache) walkByClient(fn func(k string, v *NotifyConfig)) {
	if t.client == nil {
		return
	}

	all, err := t.client.All()
	if err != nil {
		log.Printf("E! fail to read all: %v", err)
		return
	}

	for k, v := range all {
		nc, err := ParseNotifyConfig([]byte(v))
		if err != nil {
			log.Printf("E! fail to get by key %s: %v", k, err)
			continue
		}

		fn(k, nc)
	}
}

func (t *NotifyConfigCache) Walk(fn func(k string, v *NotifyConfig)) {
	t.walkByClient(fn)

	for k, v := range t.C.Items() {
		if config, ok := v.Object.(*NotifyConfig); ok {
			fn(k, config)
		}
	}
}
