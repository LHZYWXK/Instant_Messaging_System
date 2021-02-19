package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Server 结构体
type Server struct {
	IP   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	maplock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// ListenMessager 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message

		// 将msg发送给全部的在线User
		s.maplock.Lock()
		for _, client := range s.OnlineMap {
			client.C <- msg
		}
		s.maplock.Unlock()
	}
}

// BroadCast 广播消息的方法
func (s *Server) BroadCast(u *User, msg string) {
	sendMsg := "[" + u.Addr + "]" + u.Name + ":" + msg

	s.Message <- sendMsg
}

// Handler 处理当前连接的业务
func (s *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	// fmt.Println("连接建立成功")

	user := NewUser(conn, s)

	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read error:", err)
			}

			// 提取用户的消息（去除\n）
			msg := string(buf[:n-1])

			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}
	}()

	// 当前Handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器

		case <-time.After(time.Minute * 3):
			// 已经超时
			// 将当前的User强制的关闭

			user.SendMsg("你已被踢出")

			// 销毁用的资源
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出当前Handler
			return //runtime.Goexit()

		}
	}
}

//NewServer 创建一个Server
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// Start 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go s.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error:", err)
			continue
		}

		// do handler
		go s.Handler(conn)
	}
}
