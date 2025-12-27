package storage

// CallOptions 定义存储层的通用操作选项
type CallOptions struct {
	Prefix  bool  // 是否使用前缀匹配
	LeaseID int64 // 租约 ID
}

// CallOption 是操作选项的函数类型
type CallOption func(*CallOptions)

// RegistryStorage 定义服务注册中心的存储抽象接口
type RegistryStorage interface {
	// Grant 创建租约，返回租约 ID
	Grant(ttl int64, opts ...CallOption) (int64, error)
	// Revoke 撤销租约
	Revoke(leaseID int64, opts ...CallOption) error
	// Delete 删除键值
	Delete(key string, opts ...CallOption) error
	// Put 存储键值
	Put(key string, val any, opts ...CallOption) error
	// KeepAliveOnce 续约一次
	KeepAliveOnce(leaseID int64, opts ...CallOption) error
	// Watch 监听键值变化
	Watch(key string, handler func(any), opts ...CallOption) error
}
