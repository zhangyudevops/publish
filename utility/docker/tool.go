package docker

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/gogf/gf/v2/os/gtime"
	"io"
	"regexp"
	"strings"
)

type ErrorMessage struct {
	Error string
}

func PrintDockerMsg(dr io.ReadCloser) error {
	var errorMessage ErrorMessage
	bufferIOReader := bufio.NewReader(dr)
	for {
		streamBytes, err := bufferIOReader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		_ = json.Unmarshal(streamBytes, &errorMessage)
		//g.Dump(streamBytes)
		//g.Log().Print(context.Background(), string(streamBytes))
		if errorMessage.Error != "" {
			return errors.New(errorMessage.Error)
		}
	}

	return nil
}

// FindIpAddress 匹配IP
func FindIpAddress(input string) string {
	partIp := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	balance := partIp + "\\." + partIp + "\\." + partIp + "\\." + partIp
	matchMe := regexp.MustCompile(balance)
	return matchMe.FindString(input)
}

func TodayDate() string {
	date := gtime.Date()
	sDate := strings.Split(date, "-")
	fDate := strings.Join(sDate, "")
	return fDate
}
