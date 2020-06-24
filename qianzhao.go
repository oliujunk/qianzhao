package main

import (
	"log"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// Orm orm引擎
	Orm *xorm.Engine
)

// Config 设备配置信息
type Config struct {
	ID                   int `xorm:"id"`
	DeviceName           string
	DeviceCode           string
	ItemCode             string
	VID                  string `xorm:"vid"`
	SerialNumber         string
	ElementName          string
	ElementNum           string
	ElementCode          string
	Longitude            string
	Latitude             string
	Elevation            string
	SoftwareVersion      string
	IP                   string `xorm:"ip"`
	Mask                 string
	Gateway              string
	HTTPPort             int `xorm:"http_port"`
	FTPPort              int `xorm:"ftp_port"`
	CommandPort          int
	ManagementIP         string `xorm:"management_ip"`
	ManagementPort       int
	SntpIP               string `xorm:"sntp_ip"`
	Sample               int
	DeviceType           string
	ManufacturersName    string
	ManufacturersAddress string
	ManufactureDate      string
	ContactPhone         string
	ContactName          string
}

// Data 数据
type Data struct {
	Timestamp int64
	E1        int16
	E2        int16
	E3        int16
	E4        int16
	E5        int16
	E6        int16
	E7        int16
	E8        int16
	E9        int16
	E10       int16
	E11       int16
	E12       int16
	E13       int16
	E14       int16
	E15       int16
	E16       int16
}

// Element 要素列表
type Element struct {
	Index   string
	Name    string
	Unit    string
	Min     int
	Max     int
	Prec    float64
	Decimal int
}

// User 用户
type User struct {
	Username string
	Password string
	Type     int
}

func init() {

	// 日志信息添加文件名行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	// 数据库
	var err error
	Orm, err = xorm.NewEngine("sqlite3", "data.db")
	checkErrExit(err)
}

func main() {

	go CommunicationStart()

	go FileOperationStart()

	go APIServerStart()

	go CommandServerStart()

	select {}
}
