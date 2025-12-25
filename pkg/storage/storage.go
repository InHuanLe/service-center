package storage

type RegistryStorage interface {
	Grant(int64, ...any) int64
	Put(string, any, ...any) error
	Delete(string, ...any) error
	Revoke(int64, ...any) error
	KeepAliveOnce(int64, ...any) error
	Watch(string, func(any) error, ...any)
}
