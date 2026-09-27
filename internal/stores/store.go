package stores

type Store interface {
	Get(key string) (string, error)
	Set(key string, value string) error
	Del(key string)
	Keys() []string
	Rename(old string, new string) error
	Pop(key string) (string, error)
}
