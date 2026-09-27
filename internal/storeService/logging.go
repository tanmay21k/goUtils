package storeService

import (
	"log"
	"os"

	"github.com/tanmay21k/goUtils/internal/stores"
)

type Metric struct {
	GetCount    int
	SetCount    int
	DelCount    int
	RenameCount int
	PopCount    int

	GetFailureCount    int
	SetFailureCount    int
	RenameFailureCount int
	PopFailureCount    int
}

type StoreService struct {
	inner  stores.Store
	logger *log.Logger
	Metric Metric
}

func NewStoreService(inner stores.Store) *StoreService {
	return &StoreService{
		inner: inner,
		logger: log.New(
			os.Stdout,
			"[log] ",
			log.Ltime|log.Lmicroseconds,
		),
	}
}

func (ss *StoreService) Get(key string) (string, error) {
	value, err := ss.inner.Get(key)

	if err != nil {
		ss.Metric.GetFailureCount++
		return "", err
	}

	ss.Metric.GetCount++

	ss.logger.Printf("GET key=%s", key)

	return value, nil
}

func (ss *StoreService) Set(key string, value string) error {
	err := ss.inner.Set(key, value)

	if err != nil {
		ss.Metric.SetFailureCount++
		return err
	}

	ss.Metric.SetCount++

	ss.logger.Printf("SET key=%s", key)

	return nil
}

func (ss *StoreService) Del(key string) {
	ss.inner.Del(key)

	ss.Metric.DelCount++

	ss.logger.Printf("DEL key=%s", key)
}

func (ss *StoreService) Keys() []string {
	return ss.inner.Keys()
}

func (ss *StoreService) Rename(old string, new string) error {
	err := ss.inner.Rename(old, new)

	if err != nil {
		ss.Metric.RenameFailureCount++
		return err
	}

	ss.Metric.RenameCount++

	ss.logger.Printf("RENAME old=%s new=%s", old, new)

	return nil
}

func (ss *StoreService) Pop(key string) (string, error) {
	value, err := ss.inner.Pop(key)

	if err != nil {
		ss.Metric.PopFailureCount++
		return "", err
	}

	ss.Metric.PopCount++

	ss.logger.Printf("POP key=%s", key)

	return value, nil
}
