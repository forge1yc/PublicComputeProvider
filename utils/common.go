package utils

import (
	"bytes"
	"encoding/json"
	"github.com/forge1yc/clog/clog"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func MergeStdOutAndErrOut(stOut,erOut *bytes.Buffer) string {

	m := new(bytes.Buffer)

	m.Write([]byte("stdOut:\n"))
	m.Write(stOut.Bytes())
	m.Write([]byte("errOut:\n"))
	m.Write(erOut.Bytes())

	r, _ := json.Marshal(m.String())

	return string(r)
}

// PingCheck ping 检测
// @Author: hyc
// @Description:
// @Date: 2021/11/29 15:10
func PingCheck(ip string) bool {

	sysType := runtime.GOOS
	var (
		c *exec.Cmd
	)
	// windows系统
	if sysType == "windows" {
		c = exec.Command("ping", "-n", "2", ip)
		return systemPingCheck(ip, c)
	}

	// mac系统
	if sysType == "darwin" {
		c = exec.Command("ping", "-t", "2", ip)
	} else {
		c = exec.Command("ping", "-c", "2", ip)
	}

	return systemPingCheck(ip,c)
}

func systemPingCheck(ip string, c *exec.Cmd) bool {
	result, err := c.CombinedOutput()
	//clog.Info("ping result: %v",string(result))

	if err != nil {
		clog.Error("combined output %v error: %v", ip, err)
		return false
	}

	if strings.Contains(string(result), "100.0% packet loss") {
		return false
	}

	pingResult := regexp.MustCompile(`.*=(.*)(.*ms)`)
	params := pingResult.FindStringSubmatch(string(result))
	delay, err := strconv.ParseFloat(strings.TrimSpace(params[1]), 32)
	if err != nil {
		clog.Error("get delay error, will set to private ip")
		return false
	}
	if delay < 1 {
		return true
	}

	return false
}

// ConvertStrSliceToIntSlice 转换
// @Author: hyc
// @Description:
// @Date: 2021/12/13 15:59
func ConvertStrSliceToIntSlice(ss []string) []int64 {

	var is = make([]int64,0)

	for _, v := range ss {
		value ,_ := strconv.ParseInt(v,10,64)
		is = append(is,value)
	}
	return is
}

func ChangeImagePathToName(imagePath string) string {
	a := strings.Replace(imagePath,"/","_",-1)
	b := strings.Replace(a,":","_",-1)

	return b;
}


