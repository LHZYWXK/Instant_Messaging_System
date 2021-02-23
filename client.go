package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

// Client 结构体
type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前client的模式
}

// NewClient 创建一个Client
func NewClient(serverIP string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		flag:       999,
	}
	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	// 返回对象
	return client
}

// DealResponse 处理server回应的消息，直接显示到标准输出
func (c *Client) DealResponse() {
	// 一旦c.conn有数据，就直接拷贝到stout标准输出中，永久阻塞监听
	_, _ = io.Copy(os.Stdout, c.conn)
}

// menu 模式菜单
func (c *Client) menu() bool {
	var Flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	_, _ = fmt.Scanln(&Flag)

	if Flag >= 0 && Flag <= 3 {
		c.flag = Flag
		return true
	} else {
		fmt.Println(">>>请输入合法范围内的数字<<<")
		return false
	}
}

// SelectUser 查询在线用户
func (c *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error: ", err)
		return
	}
}

// PrivateChat 私聊模式
func (c *Client) PrivateChat() {
	c.SelectUser()

	var remoteName string
	var chatMsg string

	fmt.Println(">>>请输入聊天对象[用户名]，exit退出")
	_, _ = fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>请输入消息内容，exit退出")
		_, _ = fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			// 消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>请输入聊天内容，exit退出")
			_, _ = fmt.Scanln(&chatMsg)
		}

		c.SelectUser()
		fmt.Println(">>>请输入聊天对象[用户名]，exit退出")
		_, _ = fmt.Scanln(&remoteName)
	}
}

// PublicChat 公聊模式
func (c *Client) PublicChat() {
	var chatMsg string
	// 提示用户输入消息
	fmt.Println(">>>请输入聊天内容，exit退出")
	_, _ = fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送给服务器

		// 消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>请输入聊天内容，exit退出")
		_, _ = fmt.Scanln(&chatMsg)
	}
}

// UpdateName 更新用户名
func (c *Client) UpdateName() bool {
	fmt.Println(">>>请输入用户名:")
	_, _ = fmt.Scanln(&c.Name)

	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false
	}

	return true
}

// Run 启动客户端
func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {

		}

		// 根据不同的模式处理不同的业务
		switch c.flag {
		case 1:
			// 公聊模式
			c.PublicChat()
			break
		case 2:
			// 私聊模式
			c.PrivateChat()
			break
		case 3:
			// 更新用户名
			c.UpdateName()
			break
		}
	}
}

var serverIP string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器IP地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认是8888）")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println(">>> 连接服务器失败...")
		return
	}

	// 单独开启一个goroutine去处理server的回执消息
	go client.DealResponse()

	fmt.Println(">>> 连接服务器成功...")

	// 启动客户端的业务
	client.Run()
}
