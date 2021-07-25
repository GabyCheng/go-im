package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {

	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

// 用户上线的业务
func (this *User) Online() {

	// 用户上线，将用户加入到 onlineMap 中
	this.server.mapLock.Lock()
	this.server.OnLineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播用户上线消息
	this.server.BroadCast(this, "已上线")
}

// 用户下线的业务
func (this *User) Offline() {
	// 用户下线，将用户从 onlineMap 中删除
	this.server.mapLock.Lock()
	delete(this.server.OnLineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播用户下线消息
	this.server.BroadCast(this, "下线")
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
}

func (this *User) ListenMessage() {

	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}

}
