package host

import (
	"bufio"
	"crypto/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexAddress *regexp.Regexp = regexp.MustCompile(
		`^(([0-9a-f]{0,4}):)?([0-9a-f]{2}):([0-9a-f]{2})\.([0-9a-f]{1})$`,
	)
	regexNetAddress *regexp.Regexp = regexp.MustCompile(
		`(([0-9a-f]{0,4}):)?([0-9a-f]{2}):([0-9a-f]{2})\.([0-9a-f]{1})/net`,
	)
	MByte = 1024000
	GByte = 1073741824
)

// read a single int from a file
func ReadFileInt(filename string) (int, error) {
	dat, err := os.Open(filename)
	defer dat.Close()
	if err != nil {
		return -1, err
	}
	r := bufio.NewReader(dat)
	s, err := r.ReadString('\n')
	s = strings.TrimSuffix(s, "\n")
	if err != nil {
		return -1, nil
	}
	num, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}
	return num, nil
}

func DoesFileExist(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return true
}

// read a single string from a file
func ReadFileString(filename string) (string, error) {
	dat, err := os.Open(filename)
	defer dat.Close()
	if err != nil {
		return "", err
	}
	r := bufio.NewReader(dat)
	s, err := r.ReadString('\n')
	s = strings.TrimSuffix(s, "\n")
	if err != nil {
		return "", nil
	}
	return s, nil
}
func WriteFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY, os.FileMode(0755))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil

}

func GenerateMac() net.HardwareAddr {
	buf := make([]byte, 6)
	var mac net.HardwareAddr

	_, err := rand.Read(buf)
	if err != nil {
	}

	// set the local bit and set leading character to zero (unicast)
	buf[0] = (buf[0] | 2) & 0xfe

	mac = append(mac, buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])

	return mac
}
