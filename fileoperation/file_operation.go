package fileoperation

import (
	"bytes"
	"fmt"
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

	"whxph.com/qianzhao/communication"
	"whxph.com/qianzhao/database"
)

var (
	// ElementConfigArr 通道配置
	ElementConfigArr []ElementConfig

	secondFileName          string
	beforeDaySecondFileName string
	minuteFileName          string
	fiveFileName            string
	logFileName             string
	elementArr              []database.Element
	firstWriteSecond        bool
	firstWriteMinute        bool
	lastSecondFileSize      int64
	lastSecond              int
	lastMinute              int
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

	_, _ = communication.Job.AddFunc("*/1 * * * * *", writeSecondData)
	_, _ = communication.Job.AddFunc("*/1 * * * * *", writeFiveData)
	_, _ = communication.Job.AddFunc("5 */1 * * * *", writeMinuteData)
	_, _ = communication.Job.AddFunc("0 0 0 */1 * *", restart)
	_, _ = communication.Job.AddFunc("5 0 0 */1 * *", updateLastDay)
}

// GetFiles 根据后缀获取文件列表
func GetFiles(suffix string) []FileInfo {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Panicln(err)
	}
	var fileList []FileInfo
	absPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if suffix == "sec" {
		currentDayName := time.Now().Format("20060102") + ".sec"
		currentDayInfo := FileInfo{}
		for _, file := range files {
			fileName := file.Name()
			if !file.IsDir() && path.Ext(fileName) == "."+suffix {
				if strings.Contains(fileName, currentDayName) {
					currentDayInfo.Name = fileName[strings.Index(fileName, ".")+1:]
					currentDayInfo.Size += file.Size()
				} else {
					info := FileInfo{file.Name(), absPath + "/" + fileName, file.Size()}
					fileList = append(fileList, info)
				}
			}
		}
		fileList = append(fileList, currentDayInfo)
	} else {
		for _, file := range files {
			if !file.IsDir() && path.Ext(file.Name()) == "."+suffix {
				info := FileInfo{file.Name(), absPath + "/" + file.Name(), file.Size()}
				fileList = append(fileList, info)
			}
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

	contentStr = contentStr[strings.Index(contentStr, " ")+1:]
	var buffer bytes.Buffer
	buffer.WriteString(contentStr)

	buffer.WriteString(" ")
	buffer.WriteString(strconv.Itoa(number))
	buffer.WriteString(" ")
	buffer.WriteString(message)
	buffer.WriteString(" ")
	buffer.WriteString(time.Now().Format("150405"))

	newContent := AddLengthToHead(buffer)

	err := ioutil.WriteFile(logFileName, newContent.Bytes(), os.ModePerm)
	if nil != err {
		log.Println(err)
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

// ReplaceString 替换文件里面的字符串
func ReplaceString(absolutePath string, old string, new string) bool {
	content, err := ioutil.ReadFile(absolutePath)
	if err != nil {
		log.Println(err)
		return false
	}
	contentStr := string(content)

	newContentStr := strings.Replace(contentStr, old, new, 1)

	err = ioutil.WriteFile(absolutePath, []byte(newContentStr), os.ModePerm)
	if nil != err {
		log.Println(err)
		return false
	}
	return true
}

func restart() {
	for !communication.SyncRTC {
	}
	firstWriteSecond = true
	firstWriteMinute = true
	lastSecondFileSize = 0
	now := time.Now()

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Println(err)
	}

	for _, file := range files {
		fileName := file.Name()
		if !file.IsDir() && path.Ext(fileName) == ".sec" || path.Ext(fileName) == ".epd" || path.Ext(fileName) == ".log" {
			fileInfo, _ := os.Stat(fileName)
			if fileInfo.ModTime().After(now) {
				_ = os.Remove(fileName)
			}
		}
	}

	// 文件准备
	parameter := database.Parameter{}
	_, _ = database.Orm.Get(&parameter)
	_ = database.Orm.Find(&elementArr)

	numArray := strings.Split(parameter.ElementNum, "/")
	nameArray := strings.Split(parameter.ElementName, "/")
	codeArray := strings.Split(parameter.ElementCode, "/")
	ElementConfigArr = ElementConfigArr[:0]
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

	secondFileName = parameter.DeviceCode + parameter.ItemCode + now.Format("20060102") + ".sec"

	beforeDaySecondFileName = parameter.DeviceCode + parameter.ItemCode + now.Add(-time.Hour*24).Format("20060102") + ".sec"

	if !fileIsExist("00." + secondFileName) {
		writeHeader("00."+secondFileName, parameter, "02")
	}
	content, _ := ioutil.ReadFile("00." + secondFileName)
	contentStr := string(content)
	contentStr = contentStr[strings.Index(contentStr, " ")+1:]
	lastSecondFileSize += int64(len(contentStr))

	minuteFileName = parameter.DeviceCode + parameter.ItemCode + now.Format("20060102") + ".epd"
	if !fileIsExist(minuteFileName) {
		writeHeader(minuteFileName, parameter, "01")
	}

	logFileName = parameter.DeviceCode + parameter.ItemCode + now.Format("20060102") + ".log"
	if !fileIsExist(logFileName) {
		writeLogHeader(logFileName, parameter)
	}

	fiveFileName = "five"
	if !fileIsExist(fiveFileName) {
		writeFiveHeader(fiveFileName, parameter)
	}

}

func writeSecondData() {
	if firstWriteSecond {
		// 补齐数据
		now := time.Now()
		seconds := now.Hour()*3600 + now.Minute()*60 + now.Second()

		files, err := ioutil.ReadDir(".")
		if err != nil {
			log.Println(err)
		}

		var secondFileList []string
		var beforeMoreOneDaySecondFileList []string
		for _, file := range files {
			fileName := file.Name()
			if !file.IsDir() && path.Ext(fileName) == ".sec" {
				if strings.Contains(fileName, secondFileName) {
					secondFileList = append(secondFileList, fileName)
				} else if !strings.Contains(fileName, beforeDaySecondFileName) {
					nameArray := strings.Split(fileName, ".")
					if len(nameArray) >= 3 {
						beforeMoreOneDaySecondFileList = append(beforeMoreOneDaySecondFileList, fileName)
					}
				}
			}
		}
		if len(beforeMoreOneDaySecondFileList) > 0 {
			content1 := make([]byte, 0)
			for _, name := range beforeMoreOneDaySecondFileList {
				content, _ := ioutil.ReadFile(name)
				content1 = append(content1, content...)
				_ = os.Remove(name)
			}
			_ = ioutil.WriteFile(beforeMoreOneDaySecondFileList[0][3:], content1, os.ModePerm)
		}
		addRows := seconds
		lastFileRows := 0
		if len(secondFileList) > 1 {
			lastSecondFileName := secondFileList[len(secondFileList)-1]
			content, _ := ioutil.ReadFile(lastSecondFileName)
			contentStr := string(content)
			rBlank := regexp.MustCompile(" ")
			blanks := len(rBlank.FindAllStringSubmatch(contentStr, -1))
			lastFileRows = blanks / len(ElementConfigArr)
			addRows = seconds - (len(secondFileList)-2)*3600 - lastFileRows
			backAddRows := addRows
			var buffer bytes.Buffer
			lastHourStr := strings.Split(lastSecondFileName, ".")[0]
			lastHour, _ := strconv.Atoi(lastHourStr)
			if lastFileRows+addRows <= 3600 {
				for i := 0; i < addRows; i++ {
					for _, value := range ElementConfigArr {
						buffer.WriteString(" ")
						v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
						vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
						buffer.WriteString(vStr)
					}
				}

				lastSecondFile, err := os.OpenFile(lastSecondFileName, os.O_APPEND|os.O_CREATE, os.ModePerm)
				if err != nil {
					log.Println(err)
				}
				_, _ = lastSecondFile.Write(buffer.Bytes())
				_ = lastSecondFile.Close()
			} else {
				if lastFileRows < 3600 {
					if backAddRows >= 3600 {
						for i := 0; i < 3600-lastFileRows; i++ {
							for range ElementConfigArr {
								buffer.WriteString(" NULL")
							}
						}
					} else {
						for i := 0; i < 3600-lastFileRows; i++ {
							for _, value := range ElementConfigArr {
								buffer.WriteString(" ")
								v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
								vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
								buffer.WriteString(vStr)
							}
						}
					}

					lastSecondFile, err := os.OpenFile(lastSecondFileName, os.O_APPEND|os.O_CREATE, os.ModePerm)
					if err != nil {
						log.Println(err)
					}
					_, _ = lastSecondFile.Write(buffer.Bytes())
					_ = lastSecondFile.Close()
					addRows = addRows - (3600 - lastFileRows)
					buffer.Reset()
				}
				lastHour++
				for i := 0; i < addRows; i++ {
					if backAddRows >= 3600 {
						for range ElementConfigArr {
							buffer.WriteString(" NULL")
						}
					} else {
						for _, value := range ElementConfigArr {
							buffer.WriteString(" ")
							v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
							vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
							buffer.WriteString(vStr)
						}
					}

					if (i+1)%3600 == 0 || i == addRows-1 {
						name := fmt.Sprintf("%02d", lastHour) + "." + secondFileName
						err = ioutil.WriteFile(name, buffer.Bytes(), os.ModePerm)
						buffer.Reset()
						lastHour++
					}
				}
			}
		} else {
			addRows = seconds
			var buffer bytes.Buffer
			for i := 0; i < addRows; i++ {
				if addRows >= 3600 {
					for range ElementConfigArr {
						buffer.WriteString(" NULL")
					}
				} else {
					for _, value := range ElementConfigArr {
						buffer.WriteString(" ")
						v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
						vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
						buffer.WriteString(vStr)
					}
				}

				if (i+1)%3600 == 0 || i == addRows-1 {
					name := fmt.Sprintf("%02d", i/3600+1) + "." + secondFileName
					err = ioutil.WriteFile(name, buffer.Bytes(), os.ModePerm)
					buffer.Reset()
				}
			}
		}
		secondFiles, err := ioutil.ReadDir(".")
		if err != nil {
			log.Println(err)
		}
		for _, file := range secondFiles {
			name := file.Name()
			if !file.IsDir() && path.Ext(name) == ".sec" && strings.Contains(name, secondFileName) && strings.Split(name, ".")[0] != "00" {
				lastSecondFileSize += file.Size()
			}
		}

		lastSecond = now.Hour()*3600 + now.Minute()*60 + now.Second()

		updateLastDay()

		firstWriteSecond = false
	} else {
		now := time.Now()
		nowSecond := now.Hour()*3600 + now.Minute()*60 + now.Second()
		if nowSecond > lastSecond {
			currentIndex := fmt.Sprintf("%02d", time.Now().Hour()+1)
			var buffer bytes.Buffer
			for _, value := range ElementConfigArr {
				buffer.WriteString(" ")
				v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
				vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
				buffer.WriteString(vStr)
			}

			content, err := ioutil.ReadFile(currentIndex + "." + secondFileName)
			if err != nil {
				log.Println(err)
				content = make([]byte, 0)
			}
			content = append(content, buffer.Bytes()...)
			err = ioutil.WriteFile(currentIndex+"."+secondFileName, content, os.ModePerm)
			if err != nil {
				log.Panicln(err)
			}

			lastSecondFileSize += int64(buffer.Len())

			updateSecondHeader()

			lastSecond = nowSecond
		}
	}
}

func writeMinuteData() {
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
		contentStr = contentStr[strings.Index(contentStr, " ")+1:]
		buffer.WriteString(contentStr)
		if addRows >= 60 {
			for i := 0; i < addRows; i++ {
				for range ElementConfigArr {
					buffer.WriteString(" NULL")
				}
			}
		} else {
			for i := 0; i < addRows; i++ {
				for _, value := range ElementConfigArr {
					buffer.WriteString(" ")
					v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
					vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
					buffer.WriteString(vStr)
				}
			}
		}

		newContent := AddLengthToHead(buffer)

		err := ioutil.WriteFile(minuteFileName, newContent.Bytes(), os.ModePerm)
		if nil != err {
			log.Println(err)
		}

		lastMinute = now.Hour()*60 + now.Minute()
		firstWriteMinute = false
	} else {
		now := time.Now()
		nowMinute := now.Hour()*60 + now.Minute()
		if nowMinute > lastMinute {
			content, _ := ioutil.ReadFile(minuteFileName)
			contentStr := string(content)
			contentStr = contentStr[strings.Index(contentStr, " ")+1:]
			var buffer bytes.Buffer
			buffer.WriteString(contentStr)
			for _, value := range ElementConfigArr {
				buffer.WriteString(" ")
				v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
				vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
				buffer.WriteString(vStr)
			}

			newContent := AddLengthToHead(buffer)

			err := ioutil.WriteFile(minuteFileName, newContent.Bytes(), os.ModePerm)
			if nil != err {
				log.Println(err)
			}
			lastMinute = nowMinute
		}
	}

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

func writeHeader(fileName string, parameter database.Parameter, sample string) {
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

	err := ioutil.WriteFile(fileName, content.Bytes(), os.ModePerm)
	if nil != err {
		log.Println(err)
	}
}

func writeLogHeader(fileName string, parameter database.Parameter) {
	now := time.Now()
	var buffer bytes.Buffer
	buffer.WriteString(now.Format("20060102"))
	buffer.WriteString(" ")
	buffer.WriteString(parameter.DeviceCode)
	buffer.WriteString(" ")
	buffer.WriteString(parameter.ItemCode + parameter.VID + parameter.SerialNumber)

	content := AddLengthToHead(buffer)

	err := ioutil.WriteFile(fileName, content.Bytes(), os.ModePerm)
	if nil != err {
		log.Println(err)
	}
}

func writeFiveHeader(fileName string, parameter database.Parameter) {
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
			v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
			vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
			buffer.WriteString(vStr)
		}
	}

	content := AddLengthToHead(buffer)

	err := ioutil.WriteFile(fileName, content.Bytes(), os.ModePerm)
	if nil != err {
		log.Println(err)
	}
}

func writeFiveData() {
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
		v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
		vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
		buffer.WriteString(vStr)
	}

	newContent := AddLengthToHead(buffer)

	err := ioutil.WriteFile(fiveFileName, newContent.Bytes(), os.ModePerm)
	if nil != err {
		log.Println(err)
	}
}

func updateSecondHeader() {
	content, _ := ioutil.ReadFile("00." + secondFileName)
	contentStr := string(content)
	contentStr = contentStr[strings.Index(contentStr, " ")+1:]
	length := int(lastSecondFileSize)
	var buffer bytes.Buffer
	lengthStr := strconv.Itoa(length)
	lengthStrLen := len(lengthStr)
	if (float64)(length+lengthStrLen) < math.Pow10(lengthStrLen) {
		buffer.WriteString(strconv.Itoa(length+lengthStrLen+1) + " ")
	} else {
		buffer.WriteString(strconv.Itoa(length+lengthStrLen+2) + " ")
	}
	buffer.WriteString(contentStr)

	err := ioutil.WriteFile("00."+secondFileName, buffer.Bytes(), os.ModePerm)
	if nil != err {
		log.Println(err)
	}
}

func updateLastDay() {
	parameter := database.Parameter{}
	_, _ = database.Orm.Get(&parameter)
	now := time.Now()
	// 前一天分钟数据补齐
	beforeDayMinuteFileName := parameter.DeviceCode + parameter.ItemCode + now.Add(-time.Hour*24).Format("20060102") + ".epd"
	content, err := ioutil.ReadFile(beforeDayMinuteFileName)
	if err != nil {
		log.Println(err)
	} else {
		contentStr := string(content)
		rBlank := regexp.MustCompile(" ")
		blanks := len(rBlank.FindAllStringSubmatch(contentStr, -1))
		addRows := 1440 - (blanks-5)/len(ElementConfigArr) + 1
		contentStr = contentStr[strings.Index(contentStr, " ")+1:]
		var buffer bytes.Buffer
		if addRows > 0 {
			buffer.WriteString(contentStr)
			for i := 0; i < addRows; i++ {
				for _, value := range ElementConfigArr {
					buffer.WriteString(" ")
					v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
					vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
					buffer.WriteString(vStr)
				}
			}
			newContent := AddLengthToHead(buffer)
			err := ioutil.WriteFile(beforeDayMinuteFileName, newContent.Bytes(), os.ModePerm)
			if nil != err {
				log.Println(err)
			}
		}
	}

	var beforeDaySecondFileNameList []string

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Println(err)
	}
	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && path.Ext(name) == ".sec" && strings.Contains(name, beforeDaySecondFileName) {
			if name == beforeDaySecondFileName {
				return // 已合并完直接返回
			}
			beforeDaySecondFileNameList = append(beforeDaySecondFileNameList, name)
		}
	}
	if len(beforeDaySecondFileNameList) <= 0 {
		return // 不存在直接返回
	}

	content1 := make([]byte, 0)
	for _, name := range beforeDaySecondFileNameList {
		content, _ := ioutil.ReadFile(name)
		content1 = append(content1, content...)
		_ = os.Remove(name)
	}
	_ = ioutil.WriteFile(beforeDaySecondFileName, content1, os.ModePerm)

	content, err = ioutil.ReadFile(beforeDaySecondFileName)
	if err != nil {
		log.Println(err)
	} else {
		contentStr := string(content)
		rBlank := regexp.MustCompile(" ")
		blanks := len(rBlank.FindAllStringSubmatch(contentStr, -1))
		addRows := 86400 - (blanks-5)/len(ElementConfigArr) + 1
		contentStr = contentStr[strings.Index(contentStr, " ")+1:]
		var buffer bytes.Buffer
		if addRows > 0 {
			buffer.WriteString(contentStr)
			for i := 0; i < addRows; i++ {
				for _, value := range ElementConfigArr {
					buffer.WriteString(" ")
					v := float64(GetFieldName("E"+strconv.Itoa(value.ChannelIndex+1), communication.CurrentData)) * value.ChannelPrec
					vStr := fmt.Sprintf("%."+strconv.Itoa(value.ChannelDecimal)+"f", v)
					buffer.WriteString(vStr)
				}
			}
			newContent := AddLengthToHead(buffer)
			err := ioutil.WriteFile(beforeDaySecondFileName, newContent.Bytes(), os.ModePerm)
			if nil != err {
				log.Println(err)
			}
		}
	}
}
