package session

import (
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/network"
	"net"
	"sync"
)

const (
	Conn  Kind = iota + 1 // 连接SESSION
	User                  // 用户SESSION
	Group                 // 组SESSION;组播(Multicast),广播(Broadcast)可用
)

type Kind int

func (k Kind) String() string {
	switch k {
	case Conn:
		return "conn"
	case User:
		return "user"
	case Group:
		return "group"
	}

	return ""
}

type Session struct {
	rw         sync.RWMutex                     // 读写锁
	conns      map[int64]network.Conn           // 连接会话（连接ID -> network.Conn）
	users      map[int64]network.Conn           // 用户会话（用户ID -> network.Conn）
	groups     map[int64]network.Conn           // 组会话（连接ID -> network.Conn）
	groupConns map[int64]map[int64]network.Conn // 自定义分组会话（组ID -> 连接ID -> network.Conn）
}

func NewSession() *Session {
	return &Session{
		conns:      make(map[int64]network.Conn),
		users:      make(map[int64]network.Conn),
		groups:     make(map[int64]network.Conn),
		groupConns: make(map[int64]map[int64]network.Conn),
	}
}

// AddConn 添加连接
func (s *Session) AddConn(conn network.Conn) {
	s.rw.Lock()
	defer s.rw.Unlock()

	cid, uid, groups := conn.ID(), conn.UID(), conn.Groups()

	s.conns[cid] = conn

	if uid != 0 {
		s.users[uid] = conn
	}

	for group := range groups {
		m, ok := s.groupConns[group]
		if !ok {
			m = make(map[int64]network.Conn)
			s.groupConns[group] = m
		}
		m[cid] = conn
		s.groups[cid] = conn
	}
}

// RemConn 移除连接
func (s *Session) RemConn(conn network.Conn) {
	s.rw.Lock()
	defer s.rw.Unlock()

	cid, uid, groups := conn.ID(), conn.UID(), conn.Groups()

	delete(s.conns, cid)

	if uid != 0 {
		delete(s.users, uid)
	}

	for group := range groups {
		delete(s.groupConns[group], cid)
		if len(s.groupConns[group]) == 0 {
			delete(s.groupConns, group)
		}
	}

	delete(s.groups, cid)
}

// Has 是否存在会话
func (s *Session) Has(kind Kind, target int64) (ok bool, err error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	switch kind {
	case Conn:
		_, ok = s.conns[target]
	case User:
		_, ok = s.users[target]
	case Group:
		_, ok = s.groupConns[target]
	default:
		err = errors.ErrInvalidSessionKind
	}

	return
}

// Bind 绑定用户ID
func (s *Session) Bind(cid, uid int64) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	conn, err := s.conn(Conn, cid)
	if err != nil {
		return err
	}

	if oldUID := conn.UID(); oldUID != 0 {
		if uid == oldUID {
			return nil
		}
		delete(s.users, oldUID)
	}

	if oldConn, ok := s.users[uid]; ok {
		oldConn.Unbind()
	}

	conn.Bind(uid)
	s.users[uid] = conn

	return nil
}

// Unbind 解绑用户ID
func (s *Session) Unbind(uid int64) (int64, error) {
	s.rw.Lock()
	defer s.rw.Unlock()

	conn, err := s.conn(User, uid)
	if err != nil {
		return 0, err
	}

	conn.Unbind()
	delete(s.users, uid)

	return conn.ID(), nil
}

// BindGroups 绑定组
func (s *Session) BindGroups(cid int64, groups []int64) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	conn, err := s.conn(Conn, cid)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if _, ok := conn.Groups()[group]; ok {
			continue
		}
		m, ok := s.groupConns[group]
		if !ok {
			m = make(map[int64]network.Conn)
			s.groupConns[group] = m
		}
		m[cid] = conn
		s.groups[cid] = conn
		conn.BindGroup(group)
	}

	return nil
}

// UnbindGroups 解绑组
// groups 解绑某些组，不传参数表示解绑所有组
func (s *Session) UnbindGroups(cid int64, groups ...int64) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	conn, err := s.conn(Conn, cid)
	if err != nil {
		return err
	}

	f := func(group int64) {
		if _, ok := conn.Groups()[group]; ok {
			delete(s.groupConns[group], cid)
			if len(s.groupConns[group]) == 0 {
				delete(s.groupConns, group)
			}
			conn.UnbindGroup(group)
			if len(conn.Groups()) == 0 {
				delete(s.groups, cid)
			}
		}
	}

	if len(groups) == 0 {
		for group := range conn.Groups() {
			f(group)
		}
	} else {
		for _, group := range groups {
			f(group)
		}
	}

	return nil
}

// LocalIP 获取本地IP
func (s *Session) LocalIP(kind Kind, target int64) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	conn, err := s.conn(kind, target)
	if err != nil {
		return "", err
	}

	return conn.LocalIP()
}

// LocalAddr 获取本地地址
func (s *Session) LocalAddr(kind Kind, target int64) (net.Addr, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	conn, err := s.conn(kind, target)
	if err != nil {
		return nil, err
	}

	return conn.LocalAddr()
}

// RemoteIP 获取远端IP
func (s *Session) RemoteIP(kind Kind, target int64) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	conn, err := s.conn(kind, target)
	if err != nil {
		return "", err
	}

	return conn.RemoteIP()
}

// RemoteAddr 获取远端地址
func (s *Session) RemoteAddr(kind Kind, target int64) (net.Addr, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	conn, err := s.conn(kind, target)
	if err != nil {
		return nil, err
	}

	return conn.RemoteAddr()
}

// Close 关闭会话
func (s *Session) Close(kind Kind, target int64, force ...bool) error {
	s.rw.RLock()
	conn, err := s.conn(kind, target)
	s.rw.RUnlock()

	if err != nil {
		return err
	}

	return conn.Close(force...)
}

// Send 发送消息（同步）
func (s *Session) Send(kind Kind, target int64, msg []byte) error {
	s.rw.RLock()
	defer s.rw.RUnlock()

	switch kind {
	case Conn, User:
		conn, err := s.conn(kind, target)
		if err != nil {
			return err
		}

		return conn.Send(msg)

	case Group:
		m, ok := s.groupConns[target]
		if !ok {
			return errors.ErrNotFoundSession
		}
		for _, conn := range m {
			_ = conn.Send(msg)
		}
	}
	return nil
}

// Push 推送消息（异步）
func (s *Session) Push(kind Kind, target int64, msg []byte) error {
	s.rw.RLock()
	defer s.rw.RUnlock()

	switch kind {
	case Conn, User:
		conn, err := s.conn(kind, target)
		if err != nil {
			return err
		}

		return conn.Push(msg)

	case Group:
		m, ok := s.groupConns[target]
		if !ok {
			return errors.ErrNotFoundSession
		}
		for _, conn := range m {
			_ = conn.Push(msg)
		}
	}
	return nil
}

// Multicast 推送组播消息（异步）
func (s *Session) Multicast(kind Kind, targets []int64, msg []byte) (n int64, err error) {
	if len(targets) == 0 {
		return
	}

	s.rw.RLock()
	defer s.rw.RUnlock()

	if kind == Group {
		conns := make(map[int64]struct{})
		for _, target := range targets {
			m, ok := s.groupConns[target]
			if ok {
				for id, conn := range m {
					if _, ok = conns[id]; !ok {
						conns[id] = struct{}{}
						if conn.Push(msg) == nil {
							n++
						}
					}
				}
			}
		}
		return
	}

	var conns map[int64]network.Conn
	switch kind {
	case Conn:
		conns = s.conns
	case User:
		conns = s.users
	default:
		err = errors.ErrInvalidSessionKind
		return
	}

	for _, target := range targets {
		conn, ok := conns[target]
		if !ok {
			continue
		}
		if conn.Push(msg) == nil {
			n++
		}
	}

	return
}

// Broadcast 推送广播消息（异步）
func (s *Session) Broadcast(kind Kind, msg []byte) (n int64, err error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	var conns map[int64]network.Conn
	switch kind {
	case Conn:
		conns = s.conns
	case User:
		conns = s.users
	case Group:
		conns = s.groups
	default:
		err = errors.ErrInvalidSessionKind
		return
	}

	for _, conn := range conns {
		if conn.Push(msg) == nil {
			n++
		}
	}

	return
}

// Stat 统计会话总数
func (s *Session) Stat(kind Kind) (int64, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	switch kind {
	case Conn:
		return int64(len(s.conns)), nil
	case User:
		return int64(len(s.users)), nil
	case Group:
		return int64(len(s.groups)), nil
	default:
		return 0, errors.ErrInvalidSessionKind
	}
}

// 获取会话
func (s *Session) conn(kind Kind, target int64) (network.Conn, error) {
	switch kind {
	case Conn:
		conn, ok := s.conns[target]
		if !ok {
			return nil, errors.ErrNotFoundSession
		}
		return conn, nil
	case User:
		conn, ok := s.users[target]
		if !ok {
			return nil, errors.ErrNotFoundSession
		}
		return conn, nil
	default:
		return nil, errors.ErrInvalidSessionKind
	}
}
