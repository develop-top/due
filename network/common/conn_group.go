package common

import "sync"

type ConnGroup struct {
	grw   sync.RWMutex       // 组锁
	group map[int64]struct{} // 所在组
}

func NewConnGroup() *ConnGroup {
	return &ConnGroup{
		group: make(map[int64]struct{}),
	}
}

// Groups 所在组
func (c *ConnGroup) Groups() map[int64]struct{} {
	c.grw.RLock()
	defer c.grw.RUnlock()
	return c.group
}

// BindGroup 绑定组
func (c *ConnGroup) BindGroup(group int64) {
	c.grw.Lock()
	defer c.grw.Unlock()
	c.group[group] = struct{}{}
}

// UnbindGroup 解绑组
func (c *ConnGroup) UnbindGroup(group int64) {
	c.grw.Lock()
	defer c.grw.Unlock()
	delete(c.group, group)
}
