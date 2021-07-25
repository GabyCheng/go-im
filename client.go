package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	serverPort int
	Name       string
	conn       net.Conn
	flag       int // 当前 client 的模式
}

func NewCilent(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		serverPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))

	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	// 返回对象
	return client

}

// 处理 server 回应的消息， 直接显示到标准输出即可
func (client *Client) DealResponse() {

	// 一旦 client.conn 有数据， 就直接 copy 到 stdout 标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)

	// 也可以使用如下代码。
	//for {
	//	buf := make([]byte, 4096)
	//	n, err := client.conn.Read(buf)
	//	if err != nil {
	//		fmt.Println("读取消息失败 err:", err)
	//	}
	//
	//	fmt.Println(string(buf[0 : n-1]))
	//}

}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>> 请输入合法范围的数字 <<<<<")
		return false
	}

}

// 公聊模式
func (client *Client) PublicChat() {

	// 提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>> 请输入聊天内容，exit 退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送给服务器，消息不为空才发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}

}

// 查询在线用户
func (client *Client) SelectUsers() {

	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}

}

// 私聊
func (client *Client) PrivateChat() {

	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>请输入聊天对象【用户名】，exit退出：")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>请输入聊天内容，exit退出：")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			// 消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>请输入消息内容，exit退出：")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUsers()
		fmt.Println(">>>>请输入聊天对象【用户名】，exit退出：")
		fmt.Scanln(&remoteName)
	}

}

// 修改用户名
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (client *Client) Run() {

	for client.flag != 0 {

		for client.menu() != true {
		}

		// 根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			// 公聊天模式
			client.PublicChat()
			break
		case 2:
			// 私聊模式
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 9999, "设置服务器端口（默认是8888）")
}

func main() {

	// 命令行解析
	flag.Parse()

	client := NewCilent(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> 服务器连接失败...")
		return
	}

	// 单独开启一个 goroutine 去处理 server 发送过来的消息
	go client.DealResponse()

	fmt.Println(">>>>> 连接服务器成功...")

	// 启动客户端的业务
	client.Run()

}
