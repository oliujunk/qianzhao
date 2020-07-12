package communication

import (
	"io"
	"log"
	"time"

	"whxph.com/qianzhao/database"

	"github.com/jacobsa/go-serial/serial"
	"github.com/robfig/cron/v3"
)

var (
	// CurrentData 当前数据
	CurrentData database.Data

	port    io.ReadWriteCloser
	element [16]int16
)

// Start 定时读取数据
func Start() {
	// 串口配置
	options := serial.OpenOptions{
		PortName:              "COM1",
		//PortName:              "/dev/ttyUSB0",
		BaudRate:              9600,
		DataBits:              8,
		StopBits:              1,
		ParityMode:            serial.PARITY_NONE,
		RTSCTSFlowControl:     false,
		InterCharacterTimeout: 100,
		MinimumReadSize:       0,
	}

	var err error
	port, err = serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	defer port.Close()

	// 定时任务
	job := cron.New(cron.WithSeconds())
	_, _ = job.AddFunc("*/1 * * * * *", read)
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

	buf := make([]byte, 128, 128)
	timeout := time.After(time.Millisecond * 800)
	recvBuf := make([]byte, 0, 64)
	for len(recvBuf) < 37 {
		c, err := port.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		recvBuf = append(recvBuf, buf[0:c]...)
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

	CurrentData = database.Data{Timestamp: time.Now().Unix(),
		E1: element[0], E2: element[1], E3: element[2], E4: element[3], E5: element[4], E6: element[5], E7: element[6], E8: element[7],
		E9: element[8], E10: element[9], E11: element[10], E12: element[11], E13: element[12], E14: element[13], E15: element[14], E16: element[15],
	}
}
