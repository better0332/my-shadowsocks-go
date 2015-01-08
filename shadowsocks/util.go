package shadowsocks

import (
	"errors"
	"fmt"
	"os"
)

const version = "1.1.3"

var IsServer bool

func SetServer(b bool) {
	IsServer = b
	if IsServer {
		newTraffic()
	}
}

func PrintVersion() {
	fmt.Println("shadowsocks-go version", version)
}

func IsFileExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.Mode()&os.ModeType == 0 {
			return true, nil
		}
		return false, errors.New(path + " exists but is not regular file")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
