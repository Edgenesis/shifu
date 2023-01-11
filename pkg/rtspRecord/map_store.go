package rtspRecord

import (
	"encoding/gob"
	"github.com/edgenesis/shifu/pkg/logger"
	"os"
	"sync"
)

type PersistMap struct {
	mu   sync.Mutex
	m    map[string]*Device
	file *os.File
}

var store *PersistMap
var needPersist bool

func InitPersistMap(filename string) {
	store = &PersistMap{m: make(map[string]*Device)}
	needPersist = filename != ""
	if !needPersist {
		logger.Warnf("no map persistence")
		return
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logger.Errorf("can't open file: %v, won't persist the map", err)
		needPersist = false
		return
	}
	store.file = f
	if err = store.load(); err != nil {
		logger.Errorf("can't load map from file: %v", err)
		return
	}
}

func (store *PersistMap) load() error {
	if !needPersist {
		return nil
	}
	d := gob.NewDecoder(store.file)
	return d.Decode(&store.m)
}

func (store *PersistMap) save() error {
	if !needPersist {
		return nil
	}
	e := gob.NewEncoder(store.file)
	return e.Encode(store.m)
}
