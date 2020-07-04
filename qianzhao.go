package main

import (
	"log"
	"whxph.com/qianzhao/apiserver"
	"whxph.com/qianzhao/commandserver"
	"whxph.com/qianzhao/communication"
	"whxph.com/qianzhao/fileoperation"

	_ "whxph.com/qianzhao/database"
)

func init() {

	// 日志信息添加文件名行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {

	go communication.Start()

	go fileoperation.Start()

	go apiserver.Start()

	go commandserver.Start()

	select {}
}
