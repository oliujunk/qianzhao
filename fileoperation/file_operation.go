package fileoperation

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"whxph.com/qianzhao/database"

	"github.com/robfig/cron"
)

var (
	// ElementConfigArr 通道配置
	ElementConfigArr []ElementConfig

	secondFile       *os.File
	minuteFile       *os.File
	logFile          *os.File
	fiveFile         *os.File
	secondFileName   string
	minuteFileName   string
	fiveFileName     string
	logFileName      string
	elementArr       []database.Element
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

// FileInfo 返回的文件信息
type FileInfo struct {
	Name string
	Path string
	Size int64
}

// Start 文件操作
func Start() {

	restart()

	// 定时任务
	job := cron.New()
	_ = job.AddFunc("*/1 * * * * *", writeSecondData)
	_ = job.AddFunc("*/1 * * * * *", writeFiveData)
	_ = job.AddFunc("0 */1 * * * *", writeMinuteData)
	_ = job.AddFunc("0 0 0 */1 * *", restart)
	job.Start()
	defer job.Stop()
	select {}
}

// GetFiles 根据后缀获取文件列表
func GetFiles(suffix string) []FileInfo {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Println("获取文件列表失败")
	}
	var fileList []FileInfo
	absPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	for _, file := range files {
		if !file.IsDir() && path.Ext(file.Name()) == "."+suffix {
			info := FileInfo{file.Name(), absPath + "/" + file.Name(), file.Size()}
			fileList = append(fileList, info)
		}
	}
	return fileList
}

// GetFile 根据日期和后缀获取文件
func GetFile(date string, suffix string) FileInfo {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Println("获取文件列表失败")
	}
	absPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	info := FileInfo{}
	for _, file := range files {
		if !file.IsDir() && path.Ext(file.Name()) == "."+suffix && strings.Contains(file.Name(), date) {
			info.Name = file.Name()
			info.Path = absPath + "/" + file.Name()
			info.Size = file.Size()
			return info
		}
	}
	return info
}

// WriteLog 日志记录
func WriteLog(message string) {
	content, _ := ioutil.ReadFile(logFileName)
	contentStr := string(content)
	rBlank := regexp.MustCompile(" ")
	blanks := len(rBlank.FindAllStringSubmatch(contentStr, -1))
	number := (blanks - 3) / 3

	contentStr = string([]byte(contentStr)[strings.Index(contentStr, " ")+1:])
	var buffer bytes.Buffer
	buffer.WriteString(contentStr)

	buffer.WriteString(" ")
	buffer.WriteString(strconv.Itoa(number))
	buffer.WriteString(" ")
	buffer.WriteString(message)
	buffer.WriteString(" ")
	buffer.WriteString(time.Now().Format("150405"))

	_, _ = logFile.Seek(0, io.SeekStart)

	newContent := AddLengthToHead(buffer)

	_, _ = logFile.WriteString(newContent.String())

	_, _ = logFile.Seek(0, io.SeekEnd)
}

func restart() {
	firstWriteSecond = true
	firstWriteMinute = true

	// 文件准备
	var err error
	parameter := database.Parameter{}
	_, _ = database.Orm.Get(&parameter)
	_ = database.Orm.Find(&elementArr)

	numArray := strings.Split(parameter.ElementNum, "/")
	nameArray := strings.Split(parameter.ElementName, "/")
	codeArray := strings.Split(parameter.ElementCode, "/")
	for index, value := range numArray {
		if value != "100" {
			elementConfig := ElementConfig{}
			elementConfig.ChannelIndex = index
			elementConfig.ChannelNum = value
			elementConfig.ChannelName = nameArray[index]
			elementConfig.ChannelCode = codeArray[index]
			var element database.Element
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
	secondFileName = parameter.DeviceCode + parameter.ItemCode + now.Format("20060102") + ".sec"
	if fileIsExist(secondFileName) {
		secondFile, err = os.OpenFile(secondFileName, os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		_, _ = secondFile.Seek(0, io.SeekEnd)
	} else {
		secondFile, err = os.Create(secondFileName)
		if err != nil {
			log.Fatal(err)
		}
		writeHeader(secondFile, parameter, "02")
	}

	minuteFileName = parameter.DeviceCode + parameter.ItemCode + now.Format("20060102") + ".epd"
	if fileIsExist(minuteFileName) {
		minuteFile, err = os.OpenFile(minuteFileName, os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		_, _ = minuteFile.Seek(0, io.SeekEnd)
	} else {
		minuteFile, err = os.Create(minuteFileName)
		if err != nil {
			log.Fatal(err)
		}
		writeHeader(minuteFile, parameter, "01")
	}

	logFileName = parameter.DeviceCode + parameter.ItemCode + now.Format("20060102") + ".log"
	if fileIsExist(logFileName) {
		logFile, err = os.OpenFile(logFileName, os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		_, _ = logFile.Seek(0, io.SeekEnd)
	} else {
		logFile, err = os.Create(logFileName)
		if err != nil {
			log.Fatal(err)
		}
		writeLogHeader(logFile, parameter)
	}

	fiveFileName = "five"
	if fileIsExist(fiveFileName) {
		fiveFile, err = os.OpenFile(fiveFileName, os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		_, _ = fiveFile.Seek(0, io.SeekEnd)
	} else {
		fiveFile, err = os.Create(fiveFileName)
		if err != nil {
			log.Fatal(err)
		}
		writeFiveHeader(fiveFile, parameter)
	}
}

func writeSecondData() {
	data := database.Data{}
	_, _ = database.Orm.Desc("timestamp").Get(&data)

	if firstWriteSecond {
		// 补齐数据
		now := time.Now()
		seconds := now.Hour()*3600 + now.Minute()*60 + now.Second()

		content, _ := ioutil.ReadFile(secondFileName)
		contentStr := string(content)
		rBlank := regexp.MustCompile(" ")
		blanks := len(rBlank.FindAllStringSubmatch(contentStr, -1))
		addRows := seconds - (blanks-5)/len(ElementConfigArr) + 1

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
		_, _ = secondFile.Seek(0, io.SeekStart)

		newContent := AddLengthToHead(buffer)

		_, _ = secondFile.WriteString(newContent.String())

		_, _ = secondFile.Seek(0, io.SeekEnd)

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
		_, _ = secondFile.Seek(0, io.SeekStart)

		newContent := AddLengthToHead(buffer)

		_, _ = secondFile.WriteString(newContent.String())

		_, _ = secondFile.Seek(0, io.SeekEnd)
	}
}

func writeMinuteData() {
	data := database.Data{}
	_, _ = database.Orm.Desc("timestamp").Get(&data)

	if firstWriteMinute {
		// 补齐数据
		now := time.Now()
		minutes := now.Hour()*60 + now.Minute()

		content, _ := ioutil.ReadFile(minuteFileName)
		contentStr := string(content)
		rBlank := regexp.MustCompile(" ")
		blanks := len(rBlank.FindAllStringSubmatch(contentStr, -1))
		addRows := minutes - (blanks-5)/len(ElementConfigArr) + 1

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
		_, _ = minuteFile.Seek(0, io.SeekStart)

		newContent := AddLengthToHead(buffer)

		_, _ = minuteFile.WriteString(newContent.String())

		_, _ = minuteFile.Seek(0, io.SeekEnd)

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
		_, _ = minuteFile.Seek(0, io.SeekStart)

		newContent := AddLengthToHead(buffer)

		_, _ = minuteFile.WriteString(newContent.String())

		_, _ = minuteFile.Seek(0, io.SeekEnd)
	}
}

// GetFieldName 获取值
func GetFieldName(columnName string, data database.Data) int64 {
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

func writeHeader(file *os.File, parameter database.Parameter, sample string) {
	now := time.Now()
	var buffer bytes.Buffer
	buffer.WriteString(now.Format("20060102"))
	buffer.WriteString(" ")
	buffer.WriteString(parameter.DeviceCode)
	buffer.WriteString(" ")
	buffer.WriteString(parameter.ItemCode + parameter.VID + parameter.SerialNumber)
	buffer.WriteString(" ")
	buffer.WriteString(sample)
	buffer.WriteString(" ")
	buffer.WriteString(strconv.Itoa(len(ElementConfigArr)))
	for _, value := range ElementConfigArr {
		buffer.WriteString(" ")
		buffer.WriteString(value.ChannelCode)
	}

	content := AddLengthToHead(buffer)

	_, _ = file.WriteString(content.String())
}

func writeLogHeader(file *os.File, parameter database.Parameter) {
	now := time.Now()
	var buffer bytes.Buffer
	buffer.WriteString(now.Format("20060102"))
	buffer.WriteString(" ")
	buffer.WriteString(parameter.DeviceCode)
	buffer.WriteString(" ")
	buffer.WriteString(parameter.ItemCode + parameter.VID + parameter.SerialNumber)

	content := AddLengthToHead(buffer)

	_, _ = file.WriteString(content.String())
}

func writeFiveHeader(file *os.File, parameter database.Parameter) {
	data := database.Data{}
	_, _ = database.Orm.Desc("timestamp").Get(&data)

	negativeM, _ := time.ParseDuration("-5m")
	nowBefore5Minute := time.Now().Add(negativeM)
	var buffer bytes.Buffer
	buffer.WriteString(nowBefore5Minute.Format("150405"))
	buffer.WriteString(" ")
	buffer.WriteString(parameter.DeviceCode)
	buffer.WriteString(" ")
	buffer.WriteString(parameter.ItemCode + parameter.VID + parameter.SerialNumber)
	buffer.WriteString(" ")
	buffer.WriteString("02")
	buffer.WriteString(" ")
	buffer.WriteString(strconv.Itoa(len(ElementConfigArr)))
	for _, value := range ElementConfigArr {
		buffer.WriteString(" ")
		buffer.WriteString(value.ChannelCode)
	}

	for i := 0; i < 300; i++ {
		for _, value := range ElementConfigArr {
			buffer.WriteString(" ")
			v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), data)) * value.ChannelPrec
			vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
			buffer.WriteString(vStr)
		}
	}

	content := AddLengthToHead(buffer)

	_, _ = file.WriteString(content.String())
}

func writeFiveData() {
	data := database.Data{}
	_, _ = database.Orm.Desc("timestamp").Get(&data)

	parameter := database.Parameter{}
	_, _ = database.Orm.Get(&parameter)

	negativeM, _ := time.ParseDuration("-5m")
	nowBefore5Minute := time.Now().Add(negativeM)

	eleLen := len(ElementConfigArr)

	var buffer bytes.Buffer
	buffer.WriteString(nowBefore5Minute.Format("150405"))
	buffer.WriteString(" ")
	buffer.WriteString(parameter.DeviceCode)
	buffer.WriteString(" ")
	buffer.WriteString(parameter.ItemCode + parameter.VID + parameter.SerialNumber)
	buffer.WriteString(" ")
	buffer.WriteString("02")
	buffer.WriteString(" ")
	buffer.WriteString(strconv.Itoa(eleLen))
	for _, value := range ElementConfigArr {
		buffer.WriteString(" ")
		buffer.WriteString(value.ChannelCode)
	}

	content, _ := ioutil.ReadFile(fiveFileName)
	contentStr := string(content)
	for i := 0; i < 6+eleLen*2; i++ {
		contentStr = contentStr[strings.Index(contentStr, " ")+1:]
	}
	buffer.WriteString(" ")
	buffer.WriteString(contentStr)

	for _, value := range ElementConfigArr {
		buffer.WriteString(" ")
		v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), data)) * value.ChannelPrec
		vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
		buffer.WriteString(vStr)
	}

	_, _ = fiveFile.Seek(0, io.SeekStart)

	newContent := AddLengthToHead(buffer)

	_, _ = fiveFile.WriteString(newContent.String())

	_, _ = fiveFile.Seek(0, io.SeekEnd)
}
