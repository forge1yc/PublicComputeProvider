package provider

import (
	"fmt"
	"github.com/docker/docker/client"
	"os"
	"os/signal"
	"strings"
	"testing"
	"time"
)

func TestContainer(t *testing.T) {
	// 两个输出暂时没有处理，需要合成一个，目前没有输出，很是奇怪
	a := make(chan struct{})
	goldLog,pullTime,runTime,err := run("bountycloud/bounty_cloud_email:v4",1,1, []string{"aas"},a, 10)
	//goldLog,pullTime,runTime,err := run("hello-world",1,1, []string{"aas"},a)
	if err != nil {
		fmt.Printf("%+v\n",err)
	}

	fmt.Printf("pullTime %+vms\n",pullTime)
	fmt.Printf("runTime %+vms\n",runTime)

	fmt.Printf("%+v\n",goldLog)
}

func TestB(t *testing.T) {

	c := make(chan os.Signal)
	signal.Notify(c)
	defer func() {
		fmt.Println("defer...")
	}()
	fmt.Println("main 1")
	//time.Sleep(time.Hour)
	<-c
	fmt.Println("main 2")
}

func TestC(t *testing.T) {
	a := "80/tcp"
	
	c := strings.Split(a,"/")
	
	fmt.Printf("%+v\n",c)
}

func TestD(t *testing.T) {
	stopChan := make(chan struct{},0)

	go func(stopChan chan struct{}) {
		fmt.Printf("%+v\n","正在运行 10 s")
		time.Sleep(time.Second * 3)

		// 结束
		stopChan <- struct{}{}
		panic("需要终止")

	}(stopChan)

	go func(stopChan chan struct{}) {
		if _,ok := <- stopChan; ok {
			fmt.Printf("%+v\n","收到了终止信息")
		}
	}(stopChan)

	select {}
}

func TestStringTrim(t *testing.T) {
	a := "bountycloud/bounty_cloud_email:v2"
	b := "bountycloud/bounty_cloud_email"
	fmt.Printf("%+v\n",strings.TrimRight(a,":"))


	fmt.Printf("%+v\n",strings.Split(a,":")[0])
	fmt.Printf("%+v\n",strings.Split(b,":")[0])

}

func TestRemove(t *testing.T) {

	cli, _ := client.NewClientWithOpts(client.FromEnv)
	removeImages(cli, "bountycloud/bounty_cloud_email:v5")
}

