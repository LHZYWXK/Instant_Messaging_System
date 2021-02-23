package main

import (
	"net"
	"strings"
)

// User 结构体
type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// NewUser 创建一个User
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// 启动监听当前User channel消息的goroutine
	go user.ListenMessage()

	return user
}

// Online 用户的上线业务
func (u *User) Online() {

	// 用户上线，将用户加入到OnlineMap中
	u.server.capslock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.capslock.Unlock()

	// 广播当前用户上线消息
	u.server.BroadCast(u, "已上线")
}

// Offline 用户的下线业务
func (u *User) Offline() {

	// 用户下线，将用户从OnlineMap中删除
	u.server.capslock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.capslock.Unlock()

	// 广播当前用户上线消息
	u.server.BroadCast(u, "下线")
}

// SendMsg 给当前User对应的客户端发送消息
func (u *User) SendMsg(msg string) {
	_, _ = u.conn.Write([]byte(msg))
}

// DoMessage 用户处理消息的业务
func (u *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户都有哪些

		u.server.capslock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMap := "[" + user.Addr + "]" + user.Name + "在线\n"
			u.SendMsg(onlineMap)
		}
		u.server.capslock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式 rename|xxx; rename是第0个元素，xxx是第1个元素
		newName := strings.Split(msg, "|")[1]

		// 判断name是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("该用户名已存在\n")
		} else {
			u.server.capslock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.capslock.Unlock()
			u.Name = newName
			u.SendMsg("已更新用户名：" + u.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式 to|xxx|消息内容

		// 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("消息格式不正确，请使用 \"to|xxx|message\"格式\n")
			return
		}

		// 根据用户名 得到对方User对象
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("该用户名不存在\n")
			return
		}

		// 获取消息内容，通过对方的User对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(u.Name + "对你说：" + content + "\n")
	} else {
		u.server.BroadCast(u, msg)
	}
}

// ListenMessage 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C

		_, _ = u.conn.Write([]byte(msg + "\n"))
	}
}
