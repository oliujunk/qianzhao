package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron"
)

var (
	secondFile     *os.File
	minuteFile     *os.File
	secondFileName string
	minuteFileName string
	// ElementConfigArr 通道配置
	ElementConfigArr []ElementConfig
	elementArr       []Element
	firstWriteSecond bool
	firstWriteMinute bool
)

// ElementConfig 通道配置
type ElementConfig struct {
	ChannelIndex   int
	ChannelNum     string
	ChannelName    string
	ChannelCode    string
	ChannelUnit    string
	ChannelPrec    float64
	ChannelDecimal int
}

// FileOperationStart 文件操作
func FileOperationStart() {

	restart()

	// 定时任务
	job := cron.New()
	job.AddFunc("*/1 * * * * *", writeSecondData)
	job.AddFunc("0 */1 * * * *", writeMinuteData)
	job.AddFunc("0 0 0 */1 * *", restart)
	job.Start()
	defer job.Stop()
	select {}
}

func restart() {
	firstWriteSecond = true
	firstWriteMinute = true

	// 文件准备
	var err error
	config := Config{}
	Orm.Get(&config)
	Orm.Find(&elementArr)

	numArray := strings.Split(config.ElementNum, "/")
	nameArray := strings.Split(config.ElementName, "/")
	codeArray := strings.Split(config.ElementCode, "/")
	for index, value := range numArray {
		if value != "100" {
			elementConfig := ElementConfig{}
			elementConfig.ChannelIndex = index
			elementConfig.ChannelNum = value
			elementConfig.ChannelName = nameArray[index]
			elementConfig.ChannelCode = codeArray[index]
			var element Element
			for _, v := range elementArr {
				if v.Index == value {
					element = v
					break
				}
			}
			elementConfig.ChannelUnit = element.Unit
			elementConfig.ChannelPrec = element.Prec
			elementConfig.ChannelDecimal = element.Decimal
			ElementConfigArr = append(ElementConfigArr, elementConfig)
		}
	}

	now := time.Now()
	secondFileName = config.DeviceCode + config.ItemCode + now.Format("2006") + now.Format("01") + now.Format("02") + ".sec"
	if fileIsExist(secondFileName) {
		secondFile, err = os.OpenFile(secondFileName, os.O_RDWR, os.ModePerm)
		checkErrExit(err)
		secondFile.Seek(0, os.SEEK_END)
	} else {
		secondFile, err = os.Create(secondFileName)
		checkErrExit(err)
		writeHeader(secondFileName, secondFile, config, "02")
	}

	minuteFileName = config.DeviceCode + config.ItemCode + now.Format("2006") + now.Format("01") + now.Format("02") + ".epd"
	if fileIsExist(minuteFileName) {
		minuteFile, err = os.OpenFile(minuteFileName, os.O_RDWR, os.ModePerm)
		checkErrExit(err)
		minuteFile.Seek(0, os.SEEK_END)
	} else {
		minuteFile, err = os.Create(minuteFileName)
		checkErrExit(err)
		writeHeader(minuteFileName, minuteFile, config, "01")
	}
}

func writeSecondData() {
	data := Data{}
	Orm.Desc("timestamp").Get(&data)

	if firstWriteSecond {
		// 补齐数据
		now := time.Now()
		seconds := now.Hour()*3600 + now.Minute()*60 + now.Second()

		content, _ := ioutil.ReadFile(secondFileName)
		contentStr := string(content)
		rBlank := regexp.MustCompile(" ")
		blanks := len(rBlank.FindAllStringSubmatch(contentStr, -1))
		addRows := seconds - (blanks-7)/len(ElementConfigArr)

		var buffer bytes.Buffer
		contentStr = string([]byte(contentStr)[strings.Index(contentStr, " ")+1:])
		buffer.WriteString(contentStr)
		for i := 0; i < addRows; i++ {
			for _, value := range ElementConfigArr {
				buffer.WriteString(" ")
				v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), data)) * value.ChannelPrec
				vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
				buffer.WriteString(vStr)
			}
		}
		secondFile.Seek(0, os.SEEK_SET)

		newContent := AddLengthToHead(buffer)

		secondFile.WriteString(newContent.String())

		secondFile.Seek(0, os.SEEK_END)

		firstWriteSecond = false
	} else {
		content, _ := ioutil.ReadFile(secondFileName)
		contentStr := string(content)
		contentStr = string([]byte(contentStr)[strings.Index(contentStr, " ")+1:])
		var buffer bytes.Buffer
		buffer.WriteString(contentStr)
		for _, value := range ElementConfigArr {
			buffer.WriteString(" ")
			v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), data)) * value.ChannelPrec
			vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
			buffer.WriteString(vStr)
		}
		secondFile.Seek(0, os.SEEK_SET)

		newContent := AddLengthToHead(buffer)

		secondFile.WriteString(newContent.String())

		secondFile.Seek(0, os.SEEK_END)
	}
}

func writeMinuteData() {
	data := Data{}
	Orm.Desc("timestamp").Get(&data)

	if firstWriteMinute {
		// 补齐数据
		now := time.Now()
		minutes := now.Hour()*60 + now.Minute()

		content, _ := ioutil.ReadFile(minuteFileName)
		contentStr := string(content)
		rBlank := regexp.MustCompile(" ")
		blanks := len(rBlank.FindAllStringSubmatch(contentStr, -1))
		addRows := minutes - (blanks-7)/len(ElementConfigArr)

		var buffer bytes.Buffer
		contentStr = string([]byte(contentStr)[strings.Index(contentStr, " ")+1:])
		buffer.WriteString(contentStr)
		for i := 0; i < addRows; i++ {
			for _, value := range ElementConfigArr {
				buffer.WriteString(" ")
				v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), data)) * value.ChannelPrec
				vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
				buffer.WriteString(vStr)
			}
		}
		minuteFile.Seek(0, os.SEEK_SET)

		newContent := AddLengthToHead(buffer)

		minuteFile.WriteString(newContent.String())

		minuteFile.Seek(0, os.SEEK_END)

		firstWriteMinute = false
	} else {
		content, _ := ioutil.ReadFile(minuteFileName)
		contentStr := string(content)
		contentStr = string([]byte(contentStr)[strings.Index(contentStr, " ")+1:])
		var buffer bytes.Buffer
		buffer.WriteString(contentStr)
		for _, value := range ElementConfigArr {
			buffer.WriteString(" ")
			v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), data)) * value.ChannelPrec
			vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
			buffer.WriteString(vStr)
		}
		minuteFile.Seek(0, os.SEEK_SET)

		newContent := AddLengthToHead(buffer)

		minuteFile.WriteString(newContent.String())

		minuteFile.Seek(0, os.SEEK_END)
	}
}

// GetFieldName 获取值
func GetFieldName(columnName string, data Data) int64 {
	var val int64
	t := reflect.TypeOf(data)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		fmt.Println("Check type error not Struct")
		return 0
	}
	fieldNum := t.NumField()
	for i := 0; i < fieldNum; i++ {
		if strings.ToUpper(t.Field(i).Name) == strings.ToUpper(columnName) {
			v := reflect.ValueOf(data)
			val := v.FieldByName(t.Field(i).Name).Int()
			return val
		}
	}
	return val
}

// AddLengthToHead 头部添加长度
func AddLengthToHead(buffer bytes.Buffer) bytes.Buffer {
	length := len(buffer.String())
	var ret bytes.Buffer

	lengthStr := strconv.Itoa(length)
	lengthStrLen := len(lengthStr)
	if (float64)(length+lengthStrLen) < math.Pow10(lengthStrLen) {
		ret.WriteString(strconv.Itoa(length+lengthStrLen+1) + " ")
	} else {
		ret.WriteString(strconv.Itoa(length+lengthStrLen+2) + " ")
	}

	ret.Write(buffer.Bytes())
	return ret
}

func fileIsExist(filename string) bool {
	_, err := os.Stat(filename)
	if nil != err {
		return false
	}

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func writeHeader(filename string, file *os.File, config Config, sample string) {
	now := time.Now()
	var buffer bytes.Buffer
	buffer.WriteString(now.Format("2006") + now.Format("01") + now.Format("02"))
	buffer.WriteString(" ")
	buffer.WriteString(config.DeviceCode)
	buffer.WriteString(" ")
	buffer.WriteString(config.ItemCode + config.VID + config.SerialNumber)
	buffer.WriteString(" ")
	buffer.WriteString(sample)
	buffer.WriteString(" ")
	buffer.WriteString(strconv.Itoa(len(ElementConfigArr)))
	for _, value := range ElementConfigArr {
		buffer.WriteString(" ")
		buffer.WriteString(value.ChannelCode)
	}

	content := AddLengthToHead(buffer)

	file.WriteString(content.String())
}
