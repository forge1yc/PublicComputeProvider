package constants

import (
	"fmt"
	"net/http"
)

const (
	// Secret 用来标识Provider，防止恶意调用支付服务，代表只有provider能够调用支付程序
	Secret = "xxxxxxxxxxxxxxxxxxxxxxxx"
)

//Controller 开发环境 URL
//const (
//	controller          = "127.0.0.1:8000"
//	LoginUrl            = "http://" + controller + "/profiles/user_login/"
//	ProviderRegisterUrl = "http://" + controller + "/provider/register/"
//	OrgQueryUrl         = "http://" + controller + "/profiles/query_org"
//	JobBlockChainState  = "http://" + controller + "/controller/job_blockchain_status"
//	ReadyUrl            = "http://" + controller + "/provider/ready"
//	NotReadyUrl         = "http://" + controller + "/provider/not_ready"
//	AckUrl              = "http://" + controller + "/provider/job_ack?job="
//	Test                = "http://" + controller + "/profiles/test"
//
//	CommitResult = "http://" + controller + "provider/finish"
//
//	channelName   = "mychannel"
//	chaincodeName = "monitoring"
//
//	//client = docker.from_env()
//	container_name = "provider_container_1"
//)

// Controller 生产环境
const (
	controller      = "www.bountycloud.net"
	LoginUrl            = "https://" + controller + "/profiles/user_login/"
	ProviderRegisterUrl = "https://" + controller + "/provider/register/"
	OrgQueryUrl         = "https://" + controller + "/profiles/query_org"
 JobBlockChainState  = "https://" + controller + "/controller/job_blockchain_status"
	ReadyUrl            = "https://" + controller + "/provider/ready"
	NotReadyUrl         = "https://" + controller + "/provider/not_ready"
	AckUrl              = "https://" + controller + "/provider/job_ack?job="
	Test                = "https://" + controller + "/profiles/test"

	CommitResult = "https://" + controller + "provider/finish"

	channelName   = "xxxx"
	chaincodeName = "xxxx"

	//client = docker.from_env()
)

// 区块链的Host和Port
const (
	BlockChainServerHost = "xxxx"
	BlockChainServerPort = "xxxx"
)

// 背书策略类型
const (
	OneOrg = iota
	TwoOrg
	AllOrg
)

// BlockChain URL
var (
	RegisterEnrollUrl = fmt.Sprintf("http://%s:%s/users", BlockChainServerHost, BlockChainServerPort)
	// UserOrg 启动时从服务端获取
	UserOrg = ""
	OrgAll  []string

	// TimeCoefficient 用来换算真实时间的时间系数(不同规格收费不同，基础系数1，而且是弹性变化，可从Controller服务器动态获得)
	TimeCoefficient = 1.0

	// GlobalToken 带有时效性的token，运行时赋值
	GlobalToken   = ""
	ChannelName   = "xxxx"
	ChaincodeName = "xxxx"

	//peerOfOrg1 = "peer0.org1.example.com"
	//peerOfOrg2 = "peer0.org2.example.com"

	Peers []string

	// EndorsementPolices 背书策略
	EndorsementPolices = map[int][]string{
		AllOrg: Peers,
	}

	//Configs json配置，目前需要
	//Configs = new(model.Configs)

	//DEBUG 状态设置
	DEBUG = true

	// JobProperties Job属性，确认的时候需要使用
	JobProperties = make(map[string]string, 0)

	// Cookie信息
	Cookies = make([]*http.Cookie,0)
)

// LogFile 文件配置路径
const (
	LogFile    = "./configs/log.json"
	ConfigFile = "./config.json"
)

// User相关
var (
	Username string
	Password string
	Cpu      int
	Ram      int
	Ip       string
	// 目前看不需要这个，可以主动更改IP的，但是启动的时候很多不准确，监控的队列就不正常了，还是加上吧，入参需要更改，获取我去修改获取公共IP的逻辑，准确就行了
	// NeedProvidePublicService bool
)

// Msg
const (
	JobAssignmentCenterTopic   = "JOB_ASSIGNMENT_CENTER"
	ResultTooLarge             = "您好，由于您的服务产生日志过大，超过了5M，被判定为持续性服务，日志不在BountyCloud进行存储，请自行查看存储结果"
)

//func init() {
//
//	// 打开文件
//	file, _ := os.Open(ConfigFile)
//
//	// 关闭文件
//	defer file.Close()
//
//	//NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
//	decoder := json.NewDecoder(file)
//
//	//conf := new(model.Configs)
//	//Decode从输入流读取下一个json编码值并保存在v指向的值里
//	err := decoder.Decode(&Configs)
//	//fmt.Printf("current configs: %+v\n",Configs)
//	if err != nil {
//		fmt.Println("Error:", err)
//	}
//}

// 返回状态码
const (
	NoProvider = 1
	IsProvider = 2
)
