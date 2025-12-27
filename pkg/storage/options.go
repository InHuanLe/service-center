package storage

// WithPrefix 返回使用前缀匹配的选项
func WithPrefix() CallOption {
	return func(o *CallOptions) {
		o.Prefix = true
	}
}

// WithLeaseID 返回指定租约 ID 的选项
func WithLeaseID(leaseID int64) CallOption {
	return func(o *CallOptions) {
		o.LeaseID = leaseID
	}
}
