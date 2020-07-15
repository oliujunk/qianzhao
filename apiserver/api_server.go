package apiserver

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"whxph.com/qianzhao/communication"
	"whxph.com/qianzhao/utils"

	"whxph.com/qianzhao/commandserver"
	"whxph.com/qianzhao/database"
	"whxph.com/qianzhao/fileoperation"

	"github.com/gin-gonic/gin"
)

// Start api
func Start() {

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.Use(cors())

	router.GET("/api/parameter", getParameter)

	router.POST("/api/parameter", postParameter)

	router.GET("/api/property", getProperty)

	router.POST("/api/property", postProperty)

	router.GET("/api/status", getStatus)

	router.POST("/api/status", postStatus)

	router.GET("/api/data", getData)

	router.GET("/api/element", getElement)

	router.POST("/api/datetime", updatetime)

	router.GET("/api/sntpsync", sntpsync)

	router.GET("/api/reset", reset)

	router.GET("/api/reboot", reboot)

	router.GET("/api/file/second", getSecond)

	router.GET("/api/file/minute", getMinute)

	router.GET("/api/file/log", getLog)

	router.GET("/api/file/download", download)

	router.POST("/api/login", login)

	router.POST("/api/user", updateUser)

	_ = router.Run(":90")
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token,Authorization,Token")
		c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE,UPDATE")
		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func getParameter(context *gin.Context) {
	fileoperation.WriteLog("11")
	parameter := database.Parameter{}
	_, _ = database.Orm.Get(&parameter)
	context.JSON(200, parameter)
}

func postParameter(context *gin.Context) {
	fileoperation.WriteLog("07")
	oldParameter := database.Parameter{}
	_, _ = database.Orm.Get(&oldParameter)

	parameter := database.Parameter{}
	_ = context.Bind(&parameter)

	if runtime.GOOS == "linux" {
		if oldParameter.HTTPPort != parameter.HTTPPort {
			result := fileoperation.ReplaceString("/etc/nginx/nginx.conf", strconv.Itoa(oldParameter.HTTPPort), strconv.Itoa(parameter.HTTPPort))
			if !result {
				context.JSON(200, false)
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
				context.JSON(200, false)
				return
			}
		}
		if oldParameter.Gateway != parameter.Gateway {
			oldGateway := net.ParseIP(oldParameter.Gateway).String()
			newGateway := net.ParseIP(parameter.Gateway).String()
			result := fileoperation.ReplaceString("/etc/dhcpcd.conf", "static routers="+oldGateway, "static routers="+newGateway)
			if !result {
				context.JSON(200, false)
				return
			}
		}
	}

	_, _ = database.Orm.Where("id = 1").Update(&parameter)
	_, _ = database.Orm.Get(&parameter)
	context.JSON(200, parameter)
}

func getProperty(context *gin.Context) {
	fileoperation.WriteLog("13")
	property := database.Property{}
	_, _ = database.Orm.Get(&property)
	context.JSON(200, property)
}

func postProperty(context *gin.Context) {
	property := database.Property{}
	_ = context.Bind(&property)
	_, _ = database.Orm.Where("id = 1").Update(&property)
	_, _ = database.Orm.Get(&property)
	context.JSON(200, property)
}

func getStatus(context *gin.Context) {
	fileoperation.WriteLog("10")
	status := database.Status{}
	_, _ = database.Orm.Get(&status)
	context.JSON(200, status)
}

func postStatus(context *gin.Context) {
	status := database.Status{}
	_ = context.Bind(&status)
	_, _ = database.Orm.Where("id = 1").Update(&status)
	_, _ = database.Orm.Get(&status)
	context.JSON(200, status)
}

func getData(context *gin.Context) {
	//data := database.Data{}
	//_, _ = database.Orm.Desc("timestamp").Get(&data)
	context.JSON(200, communication.CurrentData)
}

func getElement(context *gin.Context) {
	var elementArr []database.Element
	_ = database.Orm.Find(&elementArr)
	context.JSON(200, elementArr)
}

type timeForm struct {
	Datetime string `form:"datetime" binding:"required"`
}

func updatetime(context *gin.Context) {
	fileoperation.WriteLog("00")
	timeForm := timeForm{}
	_ = context.Bind(&timeForm)
	if commandserver.UpdateSystemDate(timeForm.Datetime) {
		context.JSON(200, true)
	} else {
		context.JSON(200, false)
	}
}

func sntpsync(context *gin.Context) {
	fileoperation.WriteLog("00")
	parameter := database.Parameter{}

	_, _ = database.Orm.Get(&parameter)
	if commandserver.SntpSync(parameter.SntpIP) {
		context.JSON(200, true)
	} else {
		context.JSON(200, false)
	}
}

func reset(context *gin.Context) {
	if commandserver.Reset() {
		context.JSON(200, true)
	} else {
		context.JSON(200, false)
	}
}

func reboot(context *gin.Context) {
	fileoperation.WriteLog("00")
	if commandserver.Reboot() {
		context.JSON(200, true)
	} else {
		context.JSON(200, false)
	}
}

func getSecond(context *gin.Context) {
	context.JSON(200, fileoperation.GetFiles("sec"))
}

func getMinute(context *gin.Context) {
	context.JSON(200, fileoperation.GetFiles("epd"))
}

func getLog(context *gin.Context) {
	context.JSON(200, fileoperation.GetFiles("log"))
}

func download(context *gin.Context) {
	fileName := context.Query("name")
	if strings.Contains(fileName, "sec") || strings.Contains(fileName, "epd") {
		fileoperation.WriteLog("09")
	} else if strings.Contains(fileName, "log") {
		fileoperation.WriteLog("12")
	}
	context.Writer.Header().Add("Content-Disposition", "attachment; filename="+fileName)
	context.Writer.Header().Add("Content-Type", "application/octet-stream")

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Println(err)
	}
	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && strings.Contains(name, fileName) {
			content, _ := ioutil.ReadFile(name)
			_, _ = context.Writer.Write(content)
		}
	}
}

type loginForm struct {
	Username string
	Password string
}

func login(context *gin.Context) {
	loginForm := loginForm{}
	_ = context.Bind(&loginForm)
	user := database.User{}
	result, err := database.Orm.Where("username = ?", loginForm.Username).Get(&user)
	if err != nil || !result {
		log.Println(err, result)
		context.JSON(200, false)
		return
	}
	if user.Password != loginForm.Password {
		context.JSON(200, false)
		return
	}
	context.JSON(200, true)
}

type updateUserForm struct {
	Username string
	OldPassword string
	NewPassword string
}

func updateUser(context *gin.Context) {
	updateUserForm := updateUserForm{}
	_ = context.Bind(&updateUserForm)
	user := database.User{}
	result, err := database.Orm.Where("username = ?", updateUserForm.Username).Get(&user)
	if err != nil || !result {
		log.Println(err, result)
		context.JSON(200, false)
		return
	}
	if user.Password != updateUserForm.OldPassword {
		context.JSON(200, false)
		return
	}
	user.Password = updateUserForm.NewPassword
	_, err = database.Orm.Where("username = ?", user.Username).Update(&user)
	if err != nil {
		log.Println(err)
		context.JSON(200, false)
		return
	}
	context.JSON(200, true)
}