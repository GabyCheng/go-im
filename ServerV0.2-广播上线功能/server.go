package main

import (
	"fmt"
	"net"
	"sync"
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
	user := NewUser(conn)

	// 用户上线，将用户加入到 onlineMap 中
	this.mapLock.Lock()
	this.OnLineMap[user.Name] = user
	this.mapLock.Unlock()

	// 广播用户上线消息
	this.BroadCast(user, "已上线")

	// 当前 handler 阻塞
	//select {}

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
