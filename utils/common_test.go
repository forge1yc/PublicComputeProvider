package utils

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)


func TestPingCheck1(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		// TODO: Add test cases.
		{name: "1", args: args{ip: "106.11.34.5"}, want: false},
		{name: "2", args: args{ip: "39.104.50.89"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PingCheck(tt.args.ip); got != tt.want {
				t.Errorf("PingCheck() = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestCommand(t *testing.T) {

	//c := exec.Command("ping", "-t", "3", "http://www.baidu.com&#34")
	c := exec.Command("ping", "-t", "3", "106.11.34.5")
	//c := exec.Command("ping", "-t", "3", "39.104.50.89")

	result, err := c.CombinedOutput()

	if err != nil {
		fmt.Printf("%+v\n",err)
	}

	fmt.Printf("%+v\n",string(result))

	if strings.Contains(string(result),"100.0% packet loss") {
		fmt.Printf("%+v\n","100% loss")
	}

	if strings.Contains(string(result),"0.0% packet loss") {
		fmt.Printf("%+v\n","right")
	}


}

func TestCC(t *testing.T) {

	fmt.Println(fmt.Sprintf("%s","aaa"))

}

func TestTrimImagePathToName(t *testing.T) {
	type args struct {
		imagePath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "1", args: struct{ imagePath string }{imagePath: "docker/asdf:v3"}, want: "docker_asdf_v3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChangeImagePathToName(tt.args.imagePath); got != tt.want {
				t.Errorf("TrimImagePathToName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime(t *testing.T) {
	a := time.Now()

	time.Sleep(1 * time.Second)

	b := time.Since(a)

	fmt.Printf("total time %+v ms\n", b.Milliseconds())
}

/**
timer测试没有问题，这样是可以通过的
 */
func TestTimer(t *testing.T) {

	timer := time.NewTicker(2 * time.Second);

	for  {
		select {
		case <- timer.C:
			fmt.Printf("%+v\n","bingle")
		}
	}

}