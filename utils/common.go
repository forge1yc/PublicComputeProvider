package utils

import (
	"bytes"
	"encoding/json"
	"github.com/forge1yc/clog/clog"
	"os/exec"
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

	c := exec.Command("ping", "-t", "2", ip)
	//c := exec.Command("ping", "-t", "3", "39.104.50.89")

	result, err := c.CombinedOutput()

	if err != nil {
		//fmt.Printf("%+v\n",err)
		clog.Error("combined output %v error: %v",ip,err)
		return false
	}
	if strings.Contains(string(result),"100.0% packet loss") {
		//fmt.Printf("%+v\n","100% loss")
		return false
	}

	if strings.Contains(string(result),"0.0% packet loss") {
		//fmt.Printf("%+v\n","100% loss")
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


