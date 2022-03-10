package main

import (
	"ComputeProviderByGo/client"
	"ComputeProviderByGo/constants"
	"ComputeProviderByGo/model"
	"ComputeProviderByGo/provider"
	"ComputeProviderByGo/utils"
	"flag"
	"fmt"
	"github.com/forge1yc/clog/clog"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	// 需要捕获线程异常
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("err:%v",err)
		}
	}()
	// 初始化参数
	flag.StringVar(&constants.Username, "username", "default", "用户名")
	flag.StringVar(&constants.Password, "password", "default", "密码")
	flag.IntVar(&constants.Cpu, "cpu", 1, "CPU大小")
	flag.IntVar(&constants.Ram, "ram", 1, "RAM大小")
	flag.Parse()

	// 初始化日志
	if err := InitLog(constants.LogFile); err != nil {
		log.Fatal(err)
	}


	// 登录验证
	err := loginValid()
	if err != nil {
		log.Fatalf("login valid err: %v", err)
		return
	}

	// ipCheckAndPush
	ipCheckAndPush()

	// 注册cpu and ram
	registerRamAndCpu()

	// 初始化区块链相关信息
	setOrgAndTimeCoefficient()

	// 初始化MQ
	go provider.StartConsumer(constants.Cpu, constants.Ram)
}

// InitLog 初始化全局日志
func InitLog(logFile string) (err error) {

	if err = clog.SetupLogWithConf(logFile); err != nil {
		fmt.Errorf("init clog fail")
		return
	}
	clog.Info("log init success")
	return
}

func main() {
	defer clog.Close()

	//TODO 开始
	provider.Start(model.User{
		Username: constants.Username,
		Cpu:      constants.Cpu,
		Ram:      constants.Ram,
	})
	// 循环
	select {}

}

// registerRamAndCpu 注册cpu and ram
// @Author: hyc
// @Description:
// @Date: 2021/12/6 20:56
func registerRamAndCpu() {
	// 这里之所以可以是因为复用一个连接了，所以是登录状态，虽然可以复用，但是必须是这种的才可以
	re, err := client.RestyClient.R().
		SetFormData(map[string]string{
			"ram": strconv.Itoa(constants.Ram),
			"cpu": strconv.Itoa(constants.Cpu),
			"ip" : constants.Ip,
		}).
		SetCookies(constants.Cookies).
		Post(constants.ProviderRegisterUrl)

	// 这类指定了cookies理论上是可以的，经过实验，确实是可以的
	//c := resty.New()
	//re, err := c.R().
	//	SetFormData(map[string]string{
	//		"ram": strconv.Itoa(constants.Ram),
	//		"cpu": strconv.Itoa(constants.Cpu),
	//		"ip" : constants.Ip,
	//	}).
	//	// 这里只要设置了cookies就可以，复用的不靠谱，理论上就是因为复用连接了，可以右时间看看那内部实现
	//	SetCookies(constants.Cookies).
	//	Post(constants.ProviderRegisterUrl)
	if err != nil {
		log.Fatalf("register ram and cpu error: %v", err)
	}

	if strings.Contains(string(re.Body()),"success") {
		clog.Info("register ram and cpu and ip: %s", string(re.Body()))
	} else {
		log.Fatalf("register ram and cpu error, please check username or password, %v", string(re.Body()))
	}

}

// LoginValid 登录校验
// @Author: hyc
// @Description:
// @Date: 2021/12/6 20:13
func loginValid() error {
	//TODO 建立会话登录看是否是有效账户
	re, err := client.RestyClient.R().
		SetFormData(map[string]string{
			"username": constants.Username,
			"password": constants.Password,
		}).
		Post(constants.LoginUrl)
	if err != nil {
		clog.Info("login err, please check username or password, err: %v", err)
		return err
	}
	if re.StatusCode() != 200 {
		err = fmt.Errorf("login fail please check username or password")
		return err
	}
	clog.Info("login success ,welcome %+v!", constants.Username)
	constants.Cookies = re.RawResponse.Request.Cookies()
	return nil
}

// SetOrg 查询并设置org相关参数，以及一些其他相关信息，验证权限等,provider
// @Author: hyc
// @Description:
// @Date: 2021/12/6 20:10
func setOrgAndTimeCoefficient() {

	// TODO 我需要在启动的时候查一次当前客户所属的Org
	org, err := client.RestyClient.R().
		SetFormData(map[string]string{
			"username": constants.Username,
			"password": constants.Password,
		}).Post(constants.OrgQueryUrl)
	if err != nil {
		log.Fatal("can't connection to controller server")
	}

	if org.StatusCode() == 200 {
		jsonBody := utils.GetJsonBody(org.Body())
		// 验证是不是provider，不是就结束，需要用户自己去开通
		isProvider := jsonBody["is_provider"].(bool)
		if !isProvider {
			log.Fatal("you are not a provider, please check in www.bountycloud.net")
		}


		constants.UserOrg = jsonBody["org"].(string)
		clog.Info("%v",jsonBody["all_org"])
		tmpOrg := jsonBody["all_org"].([]interface{})
		for _, oo := range tmpOrg {
			constants.OrgAll = append(constants.OrgAll,oo.(string))
		}

		// 构造所有的peer节点
		for _,o := range constants.OrgAll {
			constants.Peers = append(constants.Peers,fmt.Sprintf("peer0.%s.example.com",o))
		}

		// 获得动态的计费信息，目前属于测试阶段，看看能不能接住
		constants.TimeCoefficient = jsonBody["time_coefficient"].(float64)


		clog.Debug("current org is %v, all_org is %v, peers is %v, coefficient is %v", constants.UserOrg, constants.OrgAll, constants.Peers, constants.TimeCoefficient)
	} else {
		clog.Error("err, code: %v",org.StatusCode())
		return
	}

}


// 检测是否具有公网IP，然后把Ip传送回去
// @Author: hyc
// @Description:
// @Date: 2021/11/29 13:11
func ipCheckAndPush() {

	urls := []string{
		"https://api.ipify.org?format=text",
		//"https://www.ipify.org",
		//"http://myexternalip.com",
		//"http://api.ident.me",
		//"http://whatismyipaddress.com/api",
	}

	for _, url := range urls {

		clog.Info("Getting IP address from %v",url)
		resp, err := http.Get(url)
		if err != nil {
			clog.Error("get ip error with host: %s, continue next", url)
			continue
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)

		ipt, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		clog.Info("received public IP is:%s\n", ipt)

		//TODO 检测IP是否有效
		isValid := utils.PingCheck(string(ipt))
		if !isValid {
			constants.Ip = "0.0.0.0"
			clog.Info("current public ip is inValid, set to default ip: %v", constants.Ip)
		} else {
			constants.Ip = string(ipt)
			clog.Info("current public ip is valid: %v", constants.Ip)
		}

		// 只获得，不推送，由Post请求管理
		//user := model.User{
		//	Username: constants.Username,
		//	Ip:       constants.Ip,
		//}
		//u, err := json.Marshal(user)
		//if err != nil {
		//	clog.Error("err: %v",err)
		//	return
		//}
		//err = provider.RetryCommitByMq(string(u), provider.IpCheck)
		//if err != nil {
		//	clog.Error("commit ip error: %v",err)
		//	return
		//}
	}

}
