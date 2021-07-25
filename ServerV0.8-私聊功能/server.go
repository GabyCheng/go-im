package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnLineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个 server 的接口
func NewServer(ip string, port int) *Server {

	server := &Server{
		Ip:        ip,
		Port:      port,
		OnLineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server

}

// 监听 Message 广播消息 channel 的 goroutine, 一旦有消息进来，就发送给全部在线的用户
func (this *Server) ListenMessager() {

	for {
		msg := <-this.Message
		// 将 msg 发送给全部在线的user
		this.mapLock.Lock()
		for _, cli := range this.OnLineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}

}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {

	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg

}

func (this *Server) Handler(conn net.Conn) {
	// 当前链接的业务
	user := NewUser(conn, this)

	user.Online()

	// 监听用户是否活跃的 channel
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
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息（去除 '\n'）
			msg := string(buf[:n-1])

			// 用户针对 msg 进行消息处理
			user.DoMessage(msg)

			isLive <- true
		}

	}()

	// 当前 handler 阻塞
	for {
		select {
		case <-isLive:
		// 当前用户是活跃的，应该重制定时器
		// 不做任务事情，为了激活 select，更新下面的定时器
		case <-time.After(time.Second * 100):
			// 已经超时
			// 将当前的 user 强制关闭
			user.SendMsg("你被踢了")

			// 关闭通道
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出当前的 Handler
			return

		}

	}
}

// 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))

	if err != nil {
		fmt.Println("net.Listen err :", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听 Message 的 goroutine
	go this.ListenMessager()

	for {

		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}

}
