package provider

import (
	"ComputeProviderByGo/model"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestStartConsumer(t *testing.T) {

	StartConsumer(1,1)

}

func Test_formatParameters(t *testing.T) {
	type args struct {
		cpu int
		ram int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
	}{
		// TODO: Add test cases.
		{name: "t1", args: args{
			cpu: 3,
			ram: 15,
		}, want: 4, want1: 16},
		{name: "t2", args: args{
			cpu: 0,
			ram: 3,
		}, want: 1, want1: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := formatParameters(tt.args.cpu, tt.args.ram)
			if got != tt.want {
				t.Errorf("formatParameters() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("formatParameters() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPushMsg(t *testing.T) {
	type args struct {
		messageBody string
		messageTag  string
		properties  map[string]string
	}

	job := model.Job{
		Id:        1,
		StartTime: 0,
		AckTime:   "",
		PullTime:  111,
		RunTime:   111,
		TotalTime: 123123123,
		Cost:      "",
		Finished:  false,
		Response:  "",
	}

	job_json,_ := json.Marshal(job)

	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				messageBody: string(job_json),
				messageTag:  "finished",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("%+v:%+v\n","成功发送消息",tt.args.messageBody)
			RetryCommitByMq(tt.args.messageBody,tt.args.messageTag)
		})
	}

}

func Test_ipCheckAndPush(t *testing.T) {
	tests := []struct {
		name    string
	}{
		// TODO: Add test cases.
		{name: "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get("106.11.34.5")
			if err != nil {
				panic(err)
			}
			ip, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			fmt.Printf("public ip %+v\n",ip)
		})
	}
}

func TestTimer(t *testing.T) {

	ctx,cancel := context.WithCancel(context.Background())

	go ttt1(ctx)
	go ttt2(ctx)

	time.Sleep(3 * time.Second)
	fmt.Printf("%+v\n","三秒钟过去了， 程序运行结束了, 发送信号")
	//c <- 0
	cancel()

	time.Sleep(4 * time.Second)
}

func ttt1(ctx context.Context) {

	timer1 := time.NewTicker(time.Second * 1)

	for {
		t1 := time.Now()
		fmt.Printf("t3:%+v\n",t1)
		t2 := <- timer1.C
		fmt.Printf("t4:%+v\n",t2)

		fmt.Printf("%+v\n","")

		// 这样不行，只会循环一次
		select {
		case <- ctx.Done(): // 这个应该是发送多次导致的，有一个就能接收一个，就是只有cancel之后，才有done
			fmt.Printf("%+v %v\n","我需要取消任务ttt1，因为收到了收到了信号",ctx.Err())
			return
		default:
			fmt.Printf("%+v\n", "任务还没有结束，没有收到信号，需要继续")
		}
	}

}

func ttt2(ctx context.Context) {

	timer1 := time.NewTicker(time.Second * 1)

	for {
		t1 := time.Now()
		fmt.Printf("t1:%+v\n",t1)
		t2 := <- timer1.C
		fmt.Printf("t2:%+v\n",t2)

		fmt.Printf("%+v\n","")

		// 这样不行，只会循环一次
		select {
		case <- ctx.Done():
			fmt.Printf("%+v %v\n","我需要取消任务ttt2，因为收到了收到了信号",ctx.Err())
			return
		default:
			fmt.Printf("%+v\n", "任务还没有结束，没有收到信号，需要继续")
		}
	}

}

func TestChan(t *testing.T) {
	ch := make(chan int,10)
	//go func(c chan int) {
	//	ch <-18
	//}(ch)
	//ch <- 18 // 这样会阻塞，因为是同步的，没有接收的地方
	go func() {
		time.Sleep(2 * time.Second)
		close(ch) // 关闭了，就不具备遍历属性了
	}()
	// 如果这里没有人写，就一直阻塞
	for value := range ch { // 因为关闭了，所以这类直接就过去了
		fmt.Printf("%+v\n",value)
	}


	//close(ch) //重复会引发异常
	//go func() {
		x, ok :=<-ch
		if !ok {
			fmt.Println("received: ", x)
		}
	//}()

	//close(ch) // close 就会收到消息，所以是没有问题的

	time.Sleep(2 * time.Second)
	x, ok =<-ch
	if!ok {
		// 所以如果不Ok，就是一个该 chan 类型的零值
		fmt.Println("channel closed, data invalid.", x)
	}
	

}

type aaa struct {
	q1 chan struct{}
}

func TestAAA(t *testing.T) {
	a := new(aaa)

	fmt.Printf("%+v\n",a.q1) // 这样看就是nil

	b := 2.3333


	fmt.Printf("%+v\n",strconv.FormatFloat(b,'f',0,64))
}

func TestSize(t *testing.T) {
	a := "123456"
	b := "中国是伟大的"
	fmt.Printf("%+v\n",len(a))
	fmt.Printf("%+v\n",len(b))

	if len(a) <= 5 * 1024 * 1024 {
		fmt.Printf("%+v\n","小于5m")
	}
}

func TestCount(t *testing.T) {
	a := 0
	a++
	fmt.Printf("%+v\n",a)
}

func TestTime(t *testing.T) {
	//1、时间戳转时间
	//nowUnix := time.Now().Unix() //获取当前时间戳
	var (
		nowUnix int64 = 1640603123
		afterUnix int64 = 1640605677
	)
	nowStr := unixToStr(nowUnix, "2006-01-02 15:04:05")
	afterStr := unixToStr(afterUnix, "2006-01-02 15:04:05")
	fmt.Printf("1、时间戳转时间：%d => %s \n", nowUnix, nowStr)
	fmt.Printf("2、时间戳转时间：%d => %s \n", afterUnix, afterStr)

	p := nowUnix + int64(time.Hour * 720 / 1000000000)
	fmt.Printf("%+v\n",int64(time.Hour/1000000000))
	fmt.Printf("%+v\n",int64(time.Hour * 720 / 1000000000))
	pa := unixToStr(p, "2006-01-02 15:04:05")
	fmt.Printf("3、时间戳转时间：%d => %s \n", p, pa)

	//2、时间转时间戳
	nowStr = time.Now().Format("2006/01/02 15:04:05") //根据指定的模板[ 2006/01/02 15:04:05 ]，返回时间。
	nowUnix, err := strToUnix(nowStr, "2006/01/02 15:04:05")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("2、时间转时间戳：%s => %d", nowStr, nowUnix)
}

//时间戳转时间
func unixToStr(timeUnix int64, layout string) string {
	timeStr := time.Unix(timeUnix, 0).Format(layout)
	return timeStr
}

//时间转时间戳
func strToUnix(timeStr, layout string) (int64, error) {
	local, err := time.LoadLocation("Asia/Shanghai") //设置时区
	if err != nil {
		return 0, err
	}
	tt, err := time.ParseInLocation(layout, timeStr, local)
	if err != nil {
		return 0, err
	}
	timeUnix := tt.Unix()
	return timeUnix, nil
}

func TestSelect(t *testing.T) {
	t1 := time.NewTicker(time.Second * 10)
	t2 := time.NewTicker(time.Second * 20)

	if b,ok := <- t1.C; ok {
		fmt.Printf("%+v %v\n","if logic t1 get", b)
	}

	for {
		select {
		case s := <- t1.C:
			fmt.Printf("%+v %v\n","t1 get", s)
		case <- t2.C:
			fmt.Printf("%+v\n","t2 get")
		}
		// 这个例子可以看出，不需要加default，会自己阻塞
		//fmt.Printf("%+v\n","asdfa")
	}
	
}