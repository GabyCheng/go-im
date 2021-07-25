package main

import (
	"net"
	"strings"
)

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

// 给对应的 user 客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {

	if msg == "who" {
		// 查询当前在线用户有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnLineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式： rename|张三
		newName := strings.Split(msg, "|")[1]
		// 判断 name 是否存在
		_, ok := this.server.OnLineMap[newName]
		if ok {
			this.SendMsg("当前用户名被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnLineMap, this.Name)
			this.server.OnLineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名：" + this.Name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式： to|张三|消息内容

		// 1 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不正确，请使用 \"to|张三|您好啊\"格式.\n")
			return
		}

		// 2 根据用户名，得到对方的 user 对象
		remoteUser, ok := this.server.OnLineMap[remoteName]
		if !ok {
			this.SendMsg("该用户名不存在\n")
			return
		}

		// 3 获取消息内容，通过对方的 user 对象，将消息发送出去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无消息内容，请重发\n")
			return
		}

		remoteUser.SendMsg(this.Name + "对你说：" + content)

	} else {
		this.server.BroadCast(this, msg)
	}

}

func (this *User) ListenMessage() {

	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}

}
