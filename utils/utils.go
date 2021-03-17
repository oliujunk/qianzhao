package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// IPMaskToInt 将ip格式的掩码转换为整型数字
// 如 255.255.255.0 对应的整型数字为 24
func IPMaskToInt(netmask string) (ones int, err error) {
	ipSplitArr := strings.Split(netmask, ".")
	if len(ipSplitArr) != 4 {
		return 0, fmt.Errorf("netmask:%v is not valid, pattern should like: 255.255.255.0", netmask)
	}
	ipv4MaskArr := make([]byte, 4)
	for i, value := range ipSplitArr {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("ipMaskToInt call strconv.Atoi error:[%v] string value is: [%s]", err, value)
		}
		if intValue > 255 {
			return 0, fmt.Errorf("netmask cannot greater than 255, current value is: [%s]", value)
		}
		ipv4MaskArr[i] = byte(intValue)
	}

	ones, _ = net.IPv4Mask(ipv4MaskArr[0], ipv4MaskArr[1], ipv4MaskArr[2], ipv4MaskArr[3]).Size()
	return ones, nil
}

func Crc16(data []byte, len int) uint16 {
	var crc uint16 = 0xFFFF
	var polynomial uint16 = 0xA001

	if len == 0 {
		return 0
	}

	for i := 0; i < len; i++ {
		crc ^= uint16(data[i]) & 0x00FF
		for j := 0; j < 8; j++ {
			if (crc & 0x0001) != 0 {
				crc >>= 1
				crc ^= polynomial
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}
