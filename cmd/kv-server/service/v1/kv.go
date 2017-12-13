package v1

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/moladb/go-http-example/pkg/rest"
)

const maxDataLen int = 512 * 1024

type KVService struct {
	kvs    map[string]string
	kvLock sync.RWMutex
}

func NewKVService() *KVService {
	return nil
}

func (s *KVService) GetAPIGroup() string {
	return "/v1"
}

func (s *KVService) ListHandlers() []rest.Handler {
	return []rest.Handler{
		{
			Resource:    "kv",
			Method:      "GET",
			Path:        "/kv/*key",
			HandlerFunc: getKVsHandler(s),
		},
		{
			Resource:    "kv",
			Method:      "PUT",
			Path:        "/kv/*key",
			HandlerFunc: putKVHandler(s),
		},
		{
			Resource:    "kv",
			Method:      "DELETE",
			Path:        "/kv/*key",
			HandlerFunc: deleteKVsHandler(s),
		},
	}
}

type KV struct {
	Key   string
	Value string
}

func (s *KVService) getByKey(key string) (string, bool) {
	return "", false
}

func (s *KVService) getByPrefix(prefix string) []KV {
	return []KV{}
}

func (s *KVService) getByRange(begin, end string) []KV {
	return []KV{}
}

func (s *KVService) putKV(key, value string) {
}

func (s *KVService) deleteByKey(key string) {
}

func (s *KVService) deleteByPrefix(prefix string) {
}

func (s *KVService) deleteByRange(begin, end string) {
}

func getKVsHandler(s *KVService) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func putKVHandler(s *KVService) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func deleteKVsHandler(s *KVService) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
