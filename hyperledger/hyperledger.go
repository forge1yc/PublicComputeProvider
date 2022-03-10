package hyperledger

import (
	"ComputeProviderByGo/client"
	"ComputeProviderByGo/constants"
	"ComputeProviderByGo/utils"
	"fmt"
	"github.com/forge1yc/clog/clog"
	"github.com/go-resty/resty/v2"
)

var (
	HfClient *Hyperledger
)

func init() {
	HfClient = new(Hyperledger)
}

type (

	// HyperledgerI 区块链接口
	// @Author: hyc
	// @Description: 这里不用全部实现，因为只使用了几个接口
	// @Date: 2021/11/14 12:18
	HyperledgerI interface {
		// RegisterUser 注册或者重新获取token
		RegisterUser(username, organization string)

		InvokeTransaction(token, channelName, chaincodeName, org, fcn string, endorsementPolicy int, args []string) error
	}

)

func setJsonTokenHeader(token string) map[string]string {
	return map[string]string{
		"authorization": "Bearer " + token,
		"content-type":  "application/json",
	}
}

// Hyperledger 区块链实现Job实现
// @Author: hyc
// @Description:
// @Date: 2021/11/14 12:47
type Hyperledger struct {
}

func (h Hyperledger) RegisterUser(username, organization string) (string,error) {

	clog.Debug("enroll user")
	response, err := client.RestyClient.R().SetHeaders(map[string]string{
		"content-type": "application/x-www-form-urlencoded",
	}).SetFormData(map[string]string{
		"username": username, "orgName": organization,
	}).Post(constants.RegisterEnrollUrl)

	if err != nil {
		clog.Info("enroll user, get token fail, err: %v", err)
		return "",err
	}
	jsonBody := utils.GetJsonBody(response.Body())
	return jsonBody["token"].(string),nil
}

// InvokeTransaction 交易计费，周期性调用，10分钟一次的协程调用之
// @Author: hyc
// @Description: 简易的交易调用函数
// @Date: 2021/11/26 16:15
func (h Hyperledger) InvokeTransaction(token, channelName, chaincodeName, org, fcn string, endorsementPolicy int, args []string) (*resty.Response, error) {

	url := fmt.Sprintf("http://%s:%s/channels/%s/chaincodes/%s", constants.BlockChainServerHost, constants.BlockChainServerPort, channelName, chaincodeName)

	var (
		edp []string
		ok  bool
	)

	if edp, ok = constants.EndorsementPolices[endorsementPolicy]; !ok {
		edp = constants.EndorsementPolices[constants.AllOrg]
	}

	clog.Debug("invoke channelName: %v ,chaincodeName: %v ,function: %v", channelName, chaincodeName, fcn)

	rep, err := client.RestyClient.R().SetHeaders(
		setJsonTokenHeader(token),
	).SetBody(map[string]interface{}{
		"peers": edp,
		"fcn":   fcn,
		"args":  args,
		"secret": constants.Secret,
	}).Post(url)
	if err != nil {
		clog.Error("invoke error, err: %v", err)
		return nil, err
	}

	clog.Debug("success result %v: ", string(rep.Body()))
	return rep, nil
}

// 十分钟一次的计费逻辑，不满足十分钟则忽略
// @Author: hyc
// @Description:
// @Date: 2021/12/6 14:56
//func retryCommitBCByHttp(token, channelName, chaincodeName, JobId, Time string) (*resty.Response, error) {
//
//	for i := 0; i < 3; i++ {
//		re, err := client.RestyClient.R().
//			// FIXME 这里需要改成管理员密账号密码，不能用这个提交
//			SetBasicAuth("forge2yc", "hou19941230l").
//			SetBody(map[string]interface{}{
//				"resultSum": sum,
//			}).Post(constants.CommitResult)
//
//		if err != nil {
//			clog.ListError([]string{"err"}, []interface{}{err})
//			time.Sleep(time.Duration(i) * time.Second)
//			continue
//		} else {
//			return re, nil
//		}
//	}
//
//	err := fmt.Errorf("commit result fail")
//	return nil, err
//}
