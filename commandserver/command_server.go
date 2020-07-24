package commandserver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	"whxph.com/qianzhao/communication"
	"whxph.com/qianzhao/database"
	"whxph.com/qianzhao/fileoperation"
	"whxph.com/qianzhao/utils"

	"github.com/axgle/mahonia"
	"github.com/gogf/gf/os/gproc"
)

var (
	cmdList []string
	linkMap map[net.Conn]LinkData
)

// LinkData 已连接socket
type LinkData struct {
	logined    bool
	stopChan   chan string
	closeTimer *time.Timer
	username   string
}

// UpdateSystemDate 更新系统时间
func UpdateSystemDate(dateTime string) bool {
	system := runtime.GOOS
	switch system {
	case "windows":
		_, err1 := gproc.ShellExec(`date  ` + strings.Split(dateTime, " ")[0])
		_, err2 := gproc.ShellExec(`time  ` + strings.Split(dateTime, " ")[1])
		if err1 != nil && err2 != nil {
			log.Println(err1, err2)
			return false
		}
		return true

	case "linux":
		_, err := gproc.ShellExec(`date -s  "` + dateTime + `"`)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
	return false
}

// SntpSync sntp授时
func SntpSync(ntpAddress string) bool {
	system := runtime.GOOS
	switch system {
	case "windows":
		go gproc.ShellRun("w32tm /stripchart /computer:" + ntpAddress)
		return true

	case "linux":
		go gproc.ShellRun("sudo ntpdate " + ntpAddress)
		return true
	}
	return false
}

// Reset 参数复位
func Reset() bool {
	defaultProperty := database.Property{}
	_, _ = database.Orm.Table("default_property").Get(&defaultProperty)
	_, _ = database.Orm.Table("property").Where("id = 1").Update(&defaultProperty)

	defaultParameter := database.Parameter{}
	_, _ = database.Orm.Table("default_parameter").Get(&defaultParameter)

	if runtime.GOOS == "linux" {
		oldParameter := database.Parameter{}
		_, _ = database.Orm.Get(&oldParameter)
		if oldParameter.HTTPPort != defaultParameter.HTTPPort {
			result := fileoperation.ReplaceString("/etc/nginx/nginx.conf", strconv.Itoa(oldParameter.HTTPPort), strconv.Itoa(defaultParameter.HTTPPort))
			if !result {
				return false
			}
		}
		if oldParameter.IP != defaultParameter.IP || oldParameter.Mask != defaultParameter.Mask {
			oldIP := net.ParseIP(oldParameter.IP)
			oldMask, _ := utils.IPMaskToInt(oldParameter.Mask)
			oldIPMask := oldIP.String() + "/" + strconv.Itoa(oldMask)
			newIP := net.ParseIP(defaultParameter.IP)
			newMask, _ := utils.IPMaskToInt(defaultParameter.Mask)
			newIPMask := newIP.String() + "/" + strconv.Itoa(newMask)
			result := fileoperation.ReplaceString("/etc/dhcpcd.conf", "static ip_address="+oldIPMask, "static ip_address="+newIPMask)
			if !result {
				return false
			}
		}
		if oldParameter.Gateway != defaultParameter.Gateway {
			oldGateway := net.ParseIP(oldParameter.Gateway).String()
			newGateway := net.ParseIP(defaultParameter.Gateway).String()
			result := fileoperation.ReplaceString("/etc/dhcpcd.conf", "static routers="+oldGateway, "static routers="+newGateway)
			if !result {
				return false
			}
		}
	}

	_, _ = database.Orm.Table("parameter").Where("id = 1").Update(&defaultParameter)
	return true
}

// Reboot 设备重启
func Reboot() bool {
	if runtime.GOOS == "linux" {
		go gproc.ShellRun("sudo reboot")
	}
	return true
}

// Start 命令服务
func Start() {
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

	parameter := database.Parameter{}
	_, _ = database.Orm.Get(&parameter)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(parameter.CommandPort))
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		go connHandler(conn)
	}
}

func connHandler(conn net.Conn) {
	linkData := LinkData{logined: false}
	linkData.closeTimer = time.NewTimer(time.Second * 30)
	linkMap[conn] = linkData
	buf := make([]byte, 1024)

	go func(conn net.Conn) {
		for {
			select {
			case <-linkMap[conn].closeTimer.C:
				_ = conn.Close()
			}
		}
	}(conn)

	for {
		cnt, err := conn.Read(buf)
		if cnt == 0 || err != nil {
			_ = conn.Close()
			break
		}

		command := string(buf[0:cnt])
		command = strings.Replace(command, "\r", "", -1)
		command = strings.Replace(command, "\n", "", -1)
		log.Printf(command)
		if checkCommand(command) {
			processCommand(command, conn)
		} else {
			_, _ = conn.Write([]byte("$err\n"))
		}
		linkData.closeTimer.Reset(time.Second * 30)
	}

}

func checkCommand(command string) bool {
	commandArr := strings.Split(command, "/")
	if len(commandArr) != 4 {
		return false
	}
	if commandArr[0] != "GET " || commandArr[2] != "http" || commandArr[3] != "1.1" {
		return false
	}

	commandArr[1] = strings.TrimRight(commandArr[1], " ")

	params := strings.Split(commandArr[1], "+")
	paramsLen, _ := strconv.Atoi(params[0])
	if len(commandArr[1]) != paramsLen {
		return false
	}

	parameter := database.Parameter{}
	_, _ = database.Orm.Get(&parameter)

	if params[1] != parameter.ItemCode+parameter.VID+parameter.SerialNumber {
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
	commandArr := strings.Split(command, "/")
	commandArr[1] = strings.TrimRight(commandArr[1], " ")
	params := strings.Split(commandArr[1], "+")
	if !linkMap[conn].logined && params[2] != "lin" {
		return
	}
	switch params[2] {
	case "dat":
		getMeasurement(params[3:], conn)
		break
	case "evt":
		retNAK(conn)
		break
	case "set":
		setParams(params[3:], conn)
		break
	case "rst":
		reset(conn)
		break
	case "rbt":
		reboot(conn)
		break
	case "cal":
		retNAK(conn)
		break
	case "adj":
		retNAK(conn)
		break
	case "stp":
		stop(conn)
		break
	case "rnp":
		retNAK(conn)
		break
	case "ctl":
		retNAK(conn)
		break
	case "ste":
		getState(conn)
		break
	case "log":
		getLog(params[3:], conn)
		break
	case "pmr":
		getParameter(params[3:], conn)
		break
	case "ppy":
		getProperty(conn)
		break
	case "lin":
		login(params[3:], conn)
		break
	case "rpw":
		changePassword(params[3:], conn)
		break
	default:
		break
	}
}

func login(params []string, conn net.Conn) {
	user := database.User{}
	result, err := database.Orm.Where("username = ?", params[0]).Get(&user)
	if err != nil || !result {
		_, _ = conn.Write([]byte("$nak\n"))
		return
	}

	if user.Password == params[1] {
		linkData := linkMap[conn]
		linkData.logined = true
		linkData.stopChan = make(chan string)
		linkData.username = params[0]
		linkMap[conn] = linkData
		_, _ = conn.Write([]byte("$ack\n"))
	} else {
		_, _ = conn.Write([]byte("$nak\n"))
	}
}

func getState(conn net.Conn) {
	var buffer bytes.Buffer
	now := time.Now()
	// 设备时钟
	buffer.WriteString(now.Format("20060102150405"))
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
	newContent := fileoperation.AddLengthToHead(buffer)

	_, _ = conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
	_, _ = conn.Write(newContent.Bytes())
	_, _ = conn.Write([]byte("\n"))
	_, _ = conn.Write([]byte("ack\n"))
}

func getParameter(params []string, conn net.Conn) {
	var buffer bytes.Buffer
	parameter := database.Parameter{}
	_, _ = database.Orm.Get(&parameter)
	switch params[0] {
	case "n":
		buffer.WriteString(parameter.IP)
		buffer.WriteString(" ")
		buffer.WriteString(parameter.Mask)
		buffer.WriteString(" ")
		buffer.WriteString(parameter.Gateway)
		buffer.WriteString(" ")
		buffer.WriteString("3")
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(parameter.HTTPPort))
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(parameter.FTPPort))
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(parameter.CommandPort))
		buffer.WriteString(" ")
		buffer.WriteString(parameter.ManagementIP)
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(parameter.ManagementPort))
		buffer.WriteString(" ")
		buffer.WriteString(parameter.SntpIP)

		newContent := fileoperation.AddLengthToHead(buffer)
		_, _ = conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
		_, _ = conn.Write(newContent.Bytes())
		_, _ = conn.Write([]byte("\n"))
		_, _ = conn.Write([]byte("ack\n"))
		fileoperation.WriteLog("11")
		break
	case "d":
		buffer.WriteString(parameter.DeviceCode)
		buffer.WriteString(" ")
		buffer.WriteString(parameter.ItemCode)
		buffer.WriteString(" ")
		buffer.WriteString(parameter.SerialNumber)
		buffer.WriteString(" ")
		buffer.WriteString(parameter.Longitude)
		buffer.WriteString(" ")
		buffer.WriteString(parameter.Latitude)
		buffer.WriteString(" ")
		buffer.WriteString(parameter.Elevation)

		newContent := fileoperation.AddLengthToHead(buffer)
		_, _ = conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
		_, _ = conn.Write(newContent.Bytes())
		_, _ = conn.Write([]byte("\n"))
		_, _ = conn.Write([]byte("ack\n"))
		fileoperation.WriteLog("11")
		break
	case "m":
		buffer.WriteString(fmt.Sprintf("%02d", parameter.Sample))
		buffer.WriteString(" ")
		buffer.WriteString(fmt.Sprintf("%02d", len(fileoperation.ElementConfigArr)))
		buffer.WriteString(" ")
		buffer.WriteString(fmt.Sprintf("%02d", len(fileoperation.ElementConfigArr)*2))
		for i := 0; i < len(fileoperation.ElementConfigArr); i++ {
			buffer.WriteString(" ")
			buffer.WriteString("1.0000")
			buffer.WriteString(" ")
			buffer.WriteString("0.00")
		}
		newContent := fileoperation.AddLengthToHead(buffer)
		_, _ = conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
		_, _ = conn.Write(newContent.Bytes())
		_, _ = conn.Write([]byte("\n"))
		_, _ = conn.Write([]byte("ack\n"))
		fileoperation.WriteLog("11")
		break
	case "clock":
		buffer.WriteString(time.Now().Format("20060102150405"))
		_, _ = conn.Write([]byte("$" + strconv.Itoa(len(buffer.String())) + "\n"))
		_, _ = conn.Write(buffer.Bytes())
		_, _ = conn.Write([]byte("\n"))
		_, _ = conn.Write([]byte("ack\n"))
		break
	default:
		break
	}
}

func getProperty(conn net.Conn) {
	var buffer bytes.Buffer
	property := database.Property{}
	_, _ = database.Orm.Get(&property)
	buffer.WriteString(property.DeviceName)
	buffer.WriteString(" ")
	buffer.WriteString(property.DeviceType)
	buffer.WriteString(" ")
	buffer.WriteString(property.ManufacturersName)
	buffer.WriteString(" ")
	buffer.WriteString(property.ManufacturersAddress)
	buffer.WriteString(" ")
	buffer.WriteString(property.ManufactureDate)
	buffer.WriteString(" ")
	buffer.WriteString(property.ContactPhone)
	buffer.WriteString(" ")
	buffer.WriteString(property.ContactName)
	buffer.WriteString(" ")
	buffer.WriteString(property.SoftwareVersion)

	newContent := fileoperation.AddLengthToHead(buffer)
	_, _ = conn.Write([]byte("$" + strconv.Itoa(len(newContent.String())) + "\n"))
	enc := mahonia.NewEncoder("gbk")
	_, _ = conn.Write([]byte(enc.ConvertString(newContent.String())))
	_, _ = conn.Write([]byte("\n"))
	_, _ = conn.Write([]byte("ack\n"))
	fileoperation.WriteLog("13")
}

func setParams(params []string, conn net.Conn) {
	switch params[0] {
	case "clock":
		t, err := time.Parse("20060102150405", params[1])
		if err != nil {
			log.Println(err)
		}
		if UpdateSystemDate(t.Format("2006-01-02 15:04:05")) {
			_, _ = conn.Write([]byte("ack\n"))
			fileoperation.WriteLog("00")
		}
		break
	case "n":
		temp := strings.Split(params[1], " ")
		parameter := database.Parameter{}
		parameter.IP = temp[1]
		parameter.Mask = temp[2]
		parameter.Gateway = temp[3]
		port, err := strconv.Atoi(temp[5])
		if err != nil {
			log.Println(err)
		}
		parameter.HTTPPort = port
		port, err = strconv.Atoi(temp[6])
		if err != nil {
			log.Println(err)
		}
		parameter.FTPPort = port
		port, err = strconv.Atoi(temp[7])
		if err != nil {
			log.Println(err)
		}
		parameter.CommandPort = port
		parameter.ManagementIP = temp[8]
		port, err = strconv.Atoi(temp[9])
		if err != nil {
			log.Println(err)
		}
		parameter.ManagementPort = port
		parameter.SntpIP = temp[10]

		if runtime.GOOS == "linux" {
			oldParameter := database.Parameter{}
			_, _ = database.Orm.Get(&oldParameter)
			if oldParameter.HTTPPort != parameter.HTTPPort {
				result := fileoperation.ReplaceString("/etc/nginx/nginx.conf", strconv.Itoa(oldParameter.HTTPPort), strconv.Itoa(parameter.HTTPPort))
				if !result {
					return
				}
			}
			if oldParameter.IP != parameter.IP || oldParameter.Mask != parameter.Mask {
				oldIP := net.ParseIP(oldParameter.IP)
				oldMask, _ := utils.IPMaskToInt(oldParameter.Mask)
				oldIPMask := oldIP.String() + "/" + strconv.Itoa(oldMask)
				newIP := net.ParseIP(parameter.IP)
				newMask, _ := utils.IPMaskToInt(parameter.Mask)
				newIPMask := newIP.String() + "/" + strconv.Itoa(newMask)
				result := fileoperation.ReplaceString("/etc/dhcpcd.conf", "static ip_address="+oldIPMask, "static ip_address="+newIPMask)
				if !result {
					return
				}
			}
			if oldParameter.Gateway != parameter.Gateway {
				oldGateway := net.ParseIP(oldParameter.Gateway).String()
				newGateway := net.ParseIP(parameter.Gateway).String()
				result := fileoperation.ReplaceString("/etc/dhcpcd.conf", "static routers="+oldGateway, "static routers="+newGateway)
				if !result {
					return
				}
			}
		}

		_, _ = database.Orm.Where("id = 1").Update(&parameter)
		_, _ = conn.Write([]byte("ack\n"))
		fileoperation.WriteLog("07")
		break
	case "d":
		temp := strings.Split(params[1], " ")
		parameter := database.Parameter{}
		parameter.DeviceCode = temp[1]
		parameter.ItemCode = temp[2]
		parameter.SerialNumber = temp[3]
		parameter.Longitude = temp[4]
		parameter.Latitude = temp[5]
		parameter.Elevation = temp[6]
		_, _ = database.Orm.Where("id = 1").Update(&parameter)
		_, _ = conn.Write([]byte("ack\n"))
		fileoperation.WriteLog("07")
		break
	case "m":
		temp := strings.Split(params[1], " ")
		parameter := database.Parameter{}
		sample, _ := strconv.Atoi(temp[1])
		parameter.Sample = sample
		_, _ = database.Orm.Where("id = 1").Update(&parameter)
		_, _ = conn.Write([]byte("ack\n"))
		fileoperation.WriteLog("07")
		break
	default:
		break
	}
}

func getMeasurement(params []string, conn net.Conn) {
	fileoperation.WriteLog("09")
	if params[0] == "0" { //实时数据
		linkData := linkMap[conn]
		parameter := database.Parameter{}
		_, _ = database.Orm.Get(&parameter)
		go func() {
			ticker := time.NewTicker(time.Second * 1)
			for {
				select {
				case <-ticker.C:
					var buffer bytes.Buffer
					buffer.WriteString(time.Now().Format("150405"))
					buffer.WriteString(" ")
					buffer.WriteString(parameter.DeviceCode)
					buffer.WriteString(" ")
					buffer.WriteString(parameter.ItemCode + parameter.VID + parameter.SerialNumber)
					buffer.WriteString(" ")
					buffer.WriteString("02")
					buffer.WriteString(" ")
					buffer.WriteString(strconv.Itoa(len(fileoperation.ElementConfigArr)))
					for _, value := range fileoperation.ElementConfigArr {
						buffer.WriteString(" ")
						buffer.WriteString(value.ChannelCode)
					}
					for _, value := range fileoperation.ElementConfigArr {
						buffer.WriteString(" ")
						v := float64(fileoperation.GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
						vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
						buffer.WriteString(vStr)
					}
					newContent := fileoperation.AddLengthToHead(buffer)
					_, _ = conn.Write([]byte("$" + strconv.Itoa(len(newContent.Bytes())) + "\n"))
					_, _ = conn.Write(newContent.Bytes())
					_, _ = conn.Write([]byte("\nack\n"))
					linkMap[conn].closeTimer.Reset(time.Second * 30)
				case message := <-linkData.stopChan:
					if message == "stop" {
						return
					}
				}
			}
		}()
	} else if params[0] == "5" && len(params) <= 1 { // 当前数据
		content, _ := ioutil.ReadFile("five")
		_, _ = conn.Write([]byte("$" + strconv.Itoa(len(content)) + "\n"))
		_, _ = conn.Write(content)
		_, _ = conn.Write([]byte("\nack\n"))
	} else {
		days, err := strconv.Atoi(params[0])
		if err != nil {
			_, _ = conn.Write([]byte("err\n"))
			return
		}
		if len(params) < days+1 {
			_, _ = conn.Write([]byte("err\n"))
			return
		}
		_, _ = conn.Write([]byte("$"))

		now := time.Now()
		for i := 0; i < days; i++ {
			before, _ := strconv.Atoi(params[1+i])
			if before == 0 {
				currentDay := time.Now().Format("20060102") + ".sec"
				files, err := ioutil.ReadDir(".")
				if err != nil {
					log.Println(err)
				}
				for _, file := range files {
					name := file.Name()
					if !file.IsDir() && strings.Contains(name, currentDay) {
						content, _ := ioutil.ReadFile(name)
						if strings.Split(name, ".")[0] == "00" {
							contentStr := string(content)
							_, _ = conn.Write([]byte(contentStr[0:strings.Index(contentStr, " ")] + "\n"))
							_, _ = conn.Write(content)
						} else {
							_, _ = conn.Write(content)
						}
					}
				}
				_, _ = conn.Write([]byte("\n"))
			} else {
				beforeDay := now.AddDate(0, 0, -before)
				fileInfo := fileoperation.GetFile(beforeDay.Format("20060102"), "sec")
				content, _ := ioutil.ReadFile(fileInfo.Name)
				contentStr := string(content)
				_, _ = conn.Write([]byte(contentStr[0:strings.Index(contentStr, " ")] + "\n"))
				_, _ = conn.Write(content)
				_, _ = conn.Write([]byte("\n"))
			}
			_, _ = conn.Write([]byte("ack\n"))
		}
	}
}

func stop(conn net.Conn) {
	linkData := linkMap[conn]
	linkData.stopChan <- "stop"
	_, _ = conn.Write([]byte("ack\n"))
}

func getLog(params []string, conn net.Conn) {
	days, err := strconv.Atoi(params[0])
	if err != nil {
		_, _ = conn.Write([]byte("err\n"))
		return
	}
	if len(params) < days+1 {
		_, _ = conn.Write([]byte("err\n"))
		return
	}
	_, _ = conn.Write([]byte("$"))
	now := time.Now()
	for i := 0; i < days; i++ {
		before, _ := strconv.Atoi(params[1+i])
		beforeDay := now.AddDate(0, 0, -before)
		fileInfo := fileoperation.GetFile(beforeDay.Format("20060102"), "log")
		content, _ := ioutil.ReadFile(fileInfo.Name)
		contentStr := string(content)
		_, _ = conn.Write([]byte(contentStr[0:strings.Index(contentStr, " ")] + "\n"))
		_, _ = conn.Write(content)
		_, _ = conn.Write([]byte("\n"))
	}
	_, _ = conn.Write([]byte("ack\n"))
	fileoperation.WriteLog("12")
}

func retNAK(conn net.Conn) {
	_, _ = conn.Write([]byte("nak\n"))
}

func reset(conn net.Conn) {
	if Reset() {
		_, _ = conn.Write([]byte("ack\n"))
	}
}

func reboot(conn net.Conn) {
	if Reboot() {
		_, _ = conn.Write([]byte("ack\n"))
	}
}

func changePassword(params []string, conn net.Conn) {
	linkData := linkMap[conn]
	user := database.User{}
	user.Username = linkData.username
	user.Password = params[0]
	_, err := database.Orm.Where("username = ?", user.Username).Update(&user)
	if err != nil {
		log.Println(err)
		_, _ = conn.Write([]byte("$nak\n"))
	} else {
		_, _ = conn.Write([]byte("$ack\n"))
		_ = conn.Close()
	}
}
