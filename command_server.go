package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	cmdList []string
	linkMap map[net.Conn]LinkData
)

// LinkData 已连接socket
type LinkData struct {
	logined bool
}

// CommandServerStart 命令服务
func CommandServerStart() {

	cmdList = []string{
		"dat",
		"evt",
		"set",
		"rst",
		"rbt",
		"cal",
		"adj",
		"stp",
		"rnp",
		"ctl",
		"ste",
		"log",
		"pmr",
		"ppy",
		"lin",
		"rpw",
	}

	linkMap = make(map[net.Conn]LinkData)

	listener, err := net.Listen("tcp", ":81")
	checkErr(err)

	for {
		conn, err := listener.Accept()
		checkErr(err)
		go connHandler(conn)
	}
}

func connHandler(conn net.Conn) {
	linkData := LinkData{logined: false}
	linkMap[conn] = linkData
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	buf := make([]byte, 1024)
	for {
		cnt, err := conn.Read(buf)
		if cnt == 0 || err != nil {
			conn.Close()
			break
		}

		command := string(buf[0:cnt])
		command = strings.Replace(command, "\r", "", -1)
		command = strings.Replace(command, "\n", "", -1)
		log.Printf(command)
		if checkCommand(command) {
			processCommand(command, conn)
		} else {
			conn.Write([]byte("$err\n"))
		}
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}
}

func checkCommand(command string) bool {
	commandArr := strings.Split(command, " ")
	if len(commandArr) != 3 {
		return false
	}
	if commandArr[0] != "GET" || commandArr[2] != "/http/1.1" || commandArr[1][0] != '/' {
		return false
	}

	commandArr[1] = string([]byte(commandArr[1])[1:])

	params := strings.Split(commandArr[1], "+")
	paramsLen, _ := strconv.Atoi(params[0])
	if len(commandArr[1]) != paramsLen {
		return false
	}

	config := Config{}
	Orm.Get(&config)

	if params[1] != config.ItemCode+config.VID+config.SerialNumber {
		return false
	}
	for _, v := range cmdList {
		if params[2] == v {
			return true
		}
	}

	return false
}

func processCommand(command string, conn net.Conn) {
	commandArr := strings.Split(command, " ")
	commandArr[1] = string([]byte(commandArr[1])[1:])
	params := strings.Split(commandArr[1], "+")
	if !linkMap[conn].logined && params[2] != "lin" {
		return
	}
	switch params[2] {
	case "dat":
		break
	case "evt":
		break
	case "set":
		break
	case "rst":
		break
	case "rbt":
		break
	case "cal":
		break
	case "adj":
		break
	case "stp":
		break
	case "rnp":
		break
	case "ctl":
		break
	case "ste":
		getStatus(conn)
		break
	case "log":
		break
	case "pmr":
		getParams(params[3:], conn)
		break
	case "ppy":
		getProperty(conn)
		break
	case "lin":
		login(params[3:], conn)
		break
	case "rpw":
		break
	default:
		break
	}

}

func login(params []string, conn net.Conn) {
	userList := make([]User, 0)
	Orm.Find(&userList)
	for _, user := range userList {
		if user.Username == params[0] && user.Password == params[1] {
			linkData := linkMap[conn]
			linkData.logined = true
			linkMap[conn] = linkData
			conn.Write([]byte("$ack\n"))
			return
		}
	}
	conn.Write([]byte("$nak\n"))
}

func getStatus(conn net.Conn) {
	var buffer bytes.Buffer
	now := time.Now()
	// 设备时钟
	buffer.WriteString(now.Format("2006") + now.Format("01") + now.Format("02") + now.Format("15") + now.Format("04") + now.Format("05"))
	buffer.WriteString(" ")
	// 时钟状态 “O”表示GPS授时，“1”表示SNTP授时，“2” 表示内部时钟
	buffer.WriteString("0")
	buffer.WriteString(" ")
	// 设备零点
	buffer.WriteString("0")
	buffer.WriteString(" ")
	// 直流电源状态 “O”表示正常，“1”表示异常
	buffer.WriteString("0")
	buffer.WriteString(" ")
	// 交流电源状态 “O”表示正常，“1”表示异常
	buffer.WriteString("1")
	buffer.WriteString(" ")
	// 自校准开关状态 “O”表示正常，“1”表示异常
	buffer.WriteString("0")
	buffer.WriteString(" ")
	// 调零开关状态 “O”表示正常，“1”表示异常
	buffer.WriteString("0")
	buffer.WriteString(" ")
	// 事件触发个数
	buffer.WriteString("0")
	buffer.WriteString(" ")
	// 异常告警状态
	buffer.WriteString("0")
	buffer.WriteString(" ")
	// 自定义状态
	buffer.WriteString("00")
	newContent := AddLengthToHead(buffer)

	conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
	conn.Write(newContent.Bytes())
	conn.Write([]byte("\n"))
	conn.Write([]byte("ack\n"))
}

func getParams(params []string, conn net.Conn) {
	var buffer bytes.Buffer
	config := Config{}
	Orm.Get(&config)
	switch params[0] {
	case "n":
		buffer.WriteString(config.IP)
		buffer.WriteString(" ")
		buffer.WriteString(config.Mask)
		buffer.WriteString(" ")
		buffer.WriteString(config.Gateway)
		buffer.WriteString(" ")
		buffer.WriteString("3")
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(config.HTTPPort))
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(config.FTPPort))
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(config.CommandPort))
		buffer.WriteString(" ")
		buffer.WriteString(config.ManagementIP)
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(config.ManagementPort))
		buffer.WriteString(" ")
		buffer.WriteString(config.SntpIP)

		newContent := AddLengthToHead(buffer)
		conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
		conn.Write(newContent.Bytes())
		conn.Write([]byte("\n"))
		conn.Write([]byte("ack\n"))
		break
	case "d":
		buffer.WriteString(config.DeviceCode)
		buffer.WriteString(" ")
		buffer.WriteString(config.ItemCode)
		buffer.WriteString(" ")
		buffer.WriteString(config.SerialNumber)
		buffer.WriteString(" ")
		buffer.WriteString(config.Longitude)
		buffer.WriteString(" ")
		buffer.WriteString(config.Latitude)
		buffer.WriteString(" ")
		buffer.WriteString(config.Elevation)
		newContent := AddLengthToHead(buffer)
		conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
		conn.Write(newContent.Bytes())
		conn.Write([]byte("\n"))
		conn.Write([]byte("ack\n"))
		break
	case "m":
		buffer.WriteString(fmt.Sprintf("%02d", config.Sample))
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(len(ElementConfigArr)))
		buffer.WriteString(" ")
		buffer.WriteString("0")
		newContent := AddLengthToHead(buffer)
		conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
		conn.Write(newContent.Bytes())
		conn.Write([]byte("\n"))
		conn.Write([]byte("ack\n"))
		break
	default:
		break
	}
}

func getProperty(conn net.Conn) {
	var buffer bytes.Buffer
	config := Config{}
	Orm.Get(&config)
	buffer.WriteString(config.IP)
	buffer.WriteString(" ")

	newContent := AddLengthToHead(buffer)
	conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
	conn.Write(newContent.Bytes())
	conn.Write([]byte("\n"))
	conn.Write([]byte("ack\n"))
}
