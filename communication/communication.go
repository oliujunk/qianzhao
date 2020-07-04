package communication

import (
	"log"
	"time"

	"whxph.com/qianzhao/database"

	"github.com/robfig/cron"
	"github.com/tarm/serial"
)

var (
	port    *serial.Port
	element [16]int16
)

// Start 定时读取数据
func Start() {
	// 串口配置
	// config := &serial.Config{Name: "COM8", Baud: 9600, ReadTimeout: time.Millisecond * 600}
	config := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600, ReadTimeout: time.Millisecond * 1000}
	var err error
	port, err = serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	// 定时任务
	job := cron.New()
	_ = job.AddFunc("*/1 * * * * *", read)
	job.Start()
	defer job.Stop()
	select {}
}

func read() {
	sendBuf := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x10, 0x44, 0x06}
	_, err := port.Write(sendBuf)
	if err != nil {
		log.Println(err)
	}

	recvBuf := make([]byte, 0)
	timeout := time.After(time.Millisecond * 600)
	for len(recvBuf) < 37 {
		buf := make([]byte, 37)
		_, err = port.Read(buf)
		if err != nil {
			log.Println(err)
		}
		recvBuf = append(recvBuf, buf...)
		select {
		case <-timeout:
			log.Println("read timeout")
			return
		default:
			continue
		}
	}
	if recvBuf[2] != 0x20 {
		return
	}

	for i := 0; i < 16; i++ {
		element[i] = ((int16)(recvBuf[3+i*2]) << 8) + (int16)(recvBuf[4+i*2])
	}

	data := database.Data{Timestamp: time.Now().Unix(),
		E1: element[0], E2: element[1], E3: element[2], E4: element[3], E5: element[4], E6: element[5], E7: element[6], E8: element[7],
		E9: element[8], E10: element[9], E11: element[10], E12: element[11], E13: element[12], E14: element[13], E15: element[14], E16: element[15],
	}
	_, err = database.Orm.InsertOne(&data)
	if err != nil {
		log.Println(err)
	}
}
