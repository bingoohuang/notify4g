package api

import (
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

type SnapshotService struct {
	Dir string
}

func (s *SnapshotService) InitDefault() error {
	return s.Init("./etc/snapshots")
}

func (s *SnapshotService) Init(dir string) error {
	var err error
	s.Dir, err = homedir.Expand(dir)
	if err != nil {
		return err
	}

	return os.MkdirAll(s.Dir, os.ModePerm)
}

func (s SnapshotService) Delete(file string) error {
	return os.Remove(filepath.Join(s.Dir, file))
}

func (s SnapshotService) Read(file string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(s.Dir, file))
}

func (s SnapshotService) Write(file string, content []byte) error {
	cf := filepath.Join(s.Dir, file)
	err := ioutil.WriteFile(cf, []byte(content), 0644)
	return err
}

func (s SnapshotService) Walk(fn func(file string, content []byte)) error {
	files, err := ioutil.ReadDir(s.Dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			b, err := s.Read(f.Name())
			if err != nil {
				return err
			}

			fn(f.Name(), b)
		}
	}

	return nil
}

func (s SnapshotService) CleanAll() {
	files, err := ioutil.ReadDir(s.Dir)
	if err != nil {
		logrus.Warnf("failed to read snapshot dir %v", err)
		return
	}

	for _, f := range files {
		if !f.IsDir() {
			_ = os.Remove(filepath.Join(s.Dir, f.Name()))
		}
	}
}

func (s SnapshotService) RecoverCache(c *NotifyConfigCache) error {
	return s.Walk(func(file string, content []byte) {
		ext := filepath.Ext(file)
		id := file[0 : len(file)-len(ext)]

		config, _ := ParseNotifyConfig(content)
		if config != nil {
			c.Write(id, config, false)
		}
	})
}
