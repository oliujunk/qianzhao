package communication

import (
	"github.com/gogf/gf/os/gproc"
	"io"
	"log"
	"math/rand"
	"runtime"
	"time"
	"whxph.com/qianzhao/utils"

	"whxph.com/qianzhao/database"

	"github.com/jacobsa/go-serial/serial"
	"github.com/robfig/cron/v3"
)

var (
	// CurrentData 当前数据
	CurrentData database.Data
	SyncRTC     bool

	port    io.ReadWriteCloser
	element [16]int16
	Job     *cron.Cron
)

// Start 定时读取数据
func Start() {
	// 串口配置
	options := serial.OpenOptions{
		//PortName:              "COM1",
		PortName:              "/dev/ttyUSB0",
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
		log.Fatalln(err)
	}

	// 定时任务
	Job = cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = Job.AddFunc("*/1 * * * * *", read)
	_, _ = Job.AddFunc("55 59 12 */1 * *", reSync)
	Job.Start()
}

func read() {
	sendBuf := []byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x12}
	crc := utils.Crc16(sendBuf, len(sendBuf))
	sendBuf = append(sendBuf, (byte)(crc))
	sendBuf = append(sendBuf, (byte)(crc>>8))
	_, err := port.Write(sendBuf)
	if err != nil {
		log.Println(err)
	}

	buf := make([]byte, 128, 128)
	timeout := time.After(time.Millisecond * 800)
	recvBuf := make([]byte, 0, 64)
	for len(recvBuf) < 41 {
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

	if recvBuf[2] != 0x24 {
		return
	}

	for i := 0; i < 16; i++ {
		element[i] = ((int16)(recvBuf[3+i*2]) << 8) + (int16)(recvBuf[4+i*2])
	}

	utcTimestamp := ((uint32)(recvBuf[35]) << 24) + ((uint32)(recvBuf[36]) << 16) + ((uint32)(recvBuf[37]) << 8) + (uint32)(recvBuf[38])

	utcTime := time.Unix(int64(utcTimestamp), 0).UTC()

	CurrentData = database.Data{Timestamp: time.Now().Unix(),
		E1: element[0], E2: element[1], E3: int(element[2])*10 + rand.Intn(9), E4: element[3], E5: element[4], E6: element[5], E7: element[6], E8: element[7],
		E9: element[8], E10: element[9], E11: element[10], E12: element[11], E13: element[12], E14: element[13], E15: element[14], E16: element[15],
	}

	if !SyncRTC && runtime.GOOS == "linux" {
		_, err := gproc.ShellExec(`date -s "` + utcTime.Format("2006-01-02 15:04:05") + `"`)
		if err != nil {
			log.Fatalln(err)
		}
		Job.Stop()
		Job.Start()
		SyncRTC = true
	} else if runtime.GOOS == "windows" {
		SyncRTC = true
	}
}

func reSync() {
	SyncRTC = false
}

func SetTime() {
	Job.Stop()
	cstZone := time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstZone)
	sendBuf := []byte{0x00, 0x10, 0x00, 0x20, 0x00, 0x04, 0x08,
		byte(now.Year() - 2000), byte(now.Month()), byte(now.Day()), byte(now.Hour()), byte(now.Minute()),
		0x01, 0x00, 0x01,
	}
	crc := utils.Crc16(sendBuf, len(sendBuf))
	sendBuf = append(sendBuf, (byte)(crc))
	sendBuf = append(sendBuf, (byte)(crc>>8))
	_, _ = port.Write(sendBuf)
	time.Sleep(1 * time.Second)
	_, _ = port.Write(sendBuf)
	time.Sleep(2 * time.Second)
	return
}
