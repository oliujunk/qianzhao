package database

import (
	"log"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// Orm orm引擎
	Orm *xorm.Engine
)
// Parameter 参数
type Parameter struct {
	ID             int `xorm:"id"`
	DeviceCode     string
	ItemCode       string
	VID            string `xorm:"vid"`
	SerialNumber   string
	ElementName    string
	ElementNum     string
	ElementCode    string
	Longitude      string
	Latitude       string
	Elevation      string
	IP             string `xorm:"ip"`
	Mask           string
	Gateway        string
	HTTPPort       int `xorm:"http_port"`
	FTPPort        int `xorm:"ftp_port"`
	CommandPort    int
	ManagementIP   string `xorm:"management_ip"`
	ManagementPort int
	SntpIP         string `xorm:"sntp_ip"`
	Sample         int
}

// Property 属性
type Property struct {
	ID                   int `xorm:"id"`
	DeviceName           string
	DeviceType           string
	ManufacturersName    string
	ManufacturersAddress string
	ManufactureDate      string
	ContactPhone         string
	ContactName          string
	SoftwareVersion      string
}

// Status 状态
type Status struct {
	ID              int `xorm:"id"`
	ClockState      string
	EquipmentZero   int
	DCState         string `xorm:"dc_state"`
	ACState         string `xorm:"ac_state"`
	AutoCalibration string
	ZeroSetting     string
	EventCount      int
	AbnormalState   string
	CustomState     string
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
	// 数据库
	var err error
	Orm, err = xorm.NewEngine("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}
}
