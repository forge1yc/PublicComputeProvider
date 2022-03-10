package provider

import (
	"ComputeProviderByGo/client"
	"ComputeProviderByGo/constants"
	"ComputeProviderByGo/hyperledger"
	"ComputeProviderByGo/model"
	"ComputeProviderByGo/utils"
	"context"
	"encoding/json"
	"fmt"
	mq_http_sdk "github.com/aliyunmq/mq-http-go-sdk"
	"github.com/forge1yc/clog/clog"
	"github.com/go-resty/resty/v2"
	"github.com/gogap/errors"
	"math"
	"strconv"
	"strings"
	"time"
)

// MQ基本配置
const (
	MqAk = "xxxx"
	MqSk = "xxxxxxxxxxxxx"
	MqHost     = "xxxxxxxxxxxxx"
	InstanceId = "xxxxxxxxxxxxxxx"
)

// Push消息类型
const (
	// Received 接收了消息，需要发送确认信息,包括provider_id(这个是根据名字联查出来的）
	Received = "received"
	Finished = "finished"
	IpCheck  = "ipCheck"
	State    = "state"
	Error    = "error"
	// InsufficientBalance 收到余额不足的信息需要通知controller，进行state改写
	InsufficientBalance = "insufficient balance"
)

var (
	mqClient   mq_http_sdk.MQClient
	mqProducer mq_http_sdk.MQProducer
)

func init() {
	mqClient = getMqClient()
	mqProducer = getMqProducer()
}

// 单例模式 拿到MQ client
// @Author: hyc
// @Description:
// @Date: 2021/11/29 11:00
func getMqClient() mq_http_sdk.MQClient {
	//cpu, ram = formatParameters(cpu, ram)
	// 设置HTTP协议客户端接入点，进入消息队列RocketMQ版控制台实例详情页面的接入点区域查看。
	endpoint := MqHost
	// AccessKey ID阿里云身份验证，在阿里云RAM控制台创建。
	accessKey := MqAk
	// AccessKey Secret阿里云身份验证，在阿里云RAM控制台创建。
	secretKey := MqSk

	// 消息所属的Topic，在消息队列RocketMQ版控制台创建。
	//topic := fmt.Sprintf("%dcpu_%dram", cpu, ram)
	//topic := "1cpu_1ram"

	// Topic所属的实例ID，在消息队列RocketMQ版控制台创建。
	// 若实例有命名空间，则实例ID必须传入；若实例无命名空间，则实例ID传入null空值或字符串空值。实例的命名空间可以在消息队列RocketMQ版控制台的实例详情页面查看。
	mqc := mq_http_sdk.NewAliyunMQClient(endpoint, accessKey, secretKey, "")

	return mqc
}

// 单例模式拿到producer
func getMqProducer() mq_http_sdk.MQProducer {
	instanceId := InstanceId
	//groupId := fmt.Sprintf("GID_%dCPU_%dRAM", cpu, ram)

	// 主题固定
	topic := "PROVIDER_CONTROLLER"

	mqp := mqClient.GetProducer(instanceId, topic)

	return mqp
}

// RoundUp 取整
// @Author: hyc
// @Description:
// @Date: 2021/11/16 12:13
func RoundUp(a int) int {
	if a <= 1 {
		return 1
	} else if a <= 2 {
		return 2
	} else if a <= 4 {
		return 4
	}

	if a%4 != 0 {
		return (a/4 + 1) * 4
	} else {
		return a
	}
}

// 格式化
// @Author: hyc
// @Description:
// @Date: 2021/11/16 12:14
func formatParameters(cpu, ram int) (int, int) {
	return RoundUp(cpu), RoundUp(ram)
}

// StartConsumer 启动消费者
// @Author: hyc
// @Description: 需要协程打开
// @Date: 2021/11/15 14:21
func StartConsumer(cpu, ram int) {

	topic := constants.JobAssignmentCenterTopic
	instanceId := InstanceId

	cpu, ram = formatParameters(cpu, ram)
	tag := fmt.Sprintf("%dcpu_%dram_tag", cpu, ram)
	groupId := fmt.Sprintf("GID_%dC_%dR", cpu, ram)
	if constants.Ip != "0.0.0.0" {
		tag = fmt.Sprintf("%dcpu_%dram_public_ip_tag", cpu, ram)
		// 多次映射分层，对公私有IP进行topic分层
		groupId = fmt.Sprintf("GID_%dC_%dR_PIP", cpu, ram)
	}

	mqConsumer := mqClient.GetConsumer(instanceId, topic, groupId, tag)

	// 包装起来，接收一条消息之后就立马结束consumer,处理完，再重新执行

	clog.Info("start to wait job..., monitor topic: %v", tag)
	for {
		// endChan 结尾通道，用来阻塞当前循环结束，与finishedChan不同的是当前通道
		endChan := make(chan int)
		defer close(endChan)
		// finishedChan 作业完成通道，阻碍拿下一条作业信息，只有这个有反馈，才能证明作业完成，可以去拿下一条
		finishedChan := make(chan int)
		defer close(finishedChan)
		// respChan 作业获取通道
		respChan := make(chan mq_http_sdk.ConsumeMessageResponse)
		defer close(respChan)
		// errChan 作业获取异常错误通道
		errChan := make(chan error)
		defer close(errChan)
		// stopChan 用来选择主动停止的通道
		stopChan := make(chan struct{},0)
		defer close(stopChan)

		go func() {
			select {
			case resp := <-respChan:
				{
					// 处理业务逻辑。
					clog.Info("Consume %d messages---->\n", len(resp.Messages))
					for _, v := range resp.Messages {

						// 需要专门的改一下NextConsumeTime,这个没用
						//v.NextConsumeTime = v.FirstConsumeTime + int64(time.Hour * 720 / 1000000)

						// 句柄
						var handles []string
						handles = append(handles, v.ReceiptHandle)
						clog.Info("\tMessageID: %s, PublishTime: %d, MessageTag: %s\n"+
							"\tConsumedTimes: %d, FirstConsumeTime: %d, NextConsumeTime: %d\n"+
							"\tBody: %s\n"+
							"\tProps: %s\n",
							v.MessageId, v.PublishTime, v.MessageTag, v.ConsumedTimes,
							v.FirstConsumeTime, v.NextConsumeTime, v.MessageBody, v.Properties)
						// 赋值消息的属性，含有UUID
						constants.JobProperties = v.Properties

						task := new(model.Task)
						// 默认不会反序列化错误
						_ = json.Unmarshal([]byte(v.MessageBody), &task)
						//task.Uuid = v.Properties["uuid"]

						// 先确认,防止多次消费，一个句柄确认一次
						// NextConsumeTime前若不确认消息消费成功，则消息会被重复消费。
						// 消息句柄有时间戳，同一条消息每次消费拿到的都不一样。这里面只有一条
						ackerr := mqConsumer.AckMessage(handles)
						if ackerr != nil {
							// 某些消息的句柄可能超时，会导致消息消费状态确认不成功。
							clog.Info("ack err: %s", ackerr)
							if errAckItems, ok := ackerr.(errors.ErrCode).Context()["Detail"].([]mq_http_sdk.ErrAckItem); ok {
								for _, errAckItem := range errAckItems {
									clog.Info("\tErrorHandle:%s, ErrorCode:%s, ErrorMsg:%s\n",
										errAckItem.ErrorHandle, errAckItem.ErrorCode, errAckItem.ErrorMsg)
								}
							} else {
								fmt.Println("ack err =", ackerr)
							}
							time.Sleep(time.Duration(3) * time.Second)
						} else {
							clog.Info("Ack ---->\n\t%s\n", handles)
						}

						ackJob := &model.AckJob{
							JobId:    task.JobId,
							Username: constants.Username,
						}
						ackJobStr, _ := json.Marshal(ackJob)
						err := RetryCommitByMq(string(ackJobStr),Received)
						if err != nil {
							// 需要直接终止，目前上层捕获了，会终止程序，因为这一步出现问题很小，到时候手动处理
							panic(fmt.Errorf("bountyCloud received job ack msg error %v, please contact bountycloud@163.com", err))
						}

						ctx, cancel := context.WithCancel(context.Background())

						// 检查区块链中的作业账本创建状态，为周期支付(按量付费)准备
						if err = checkWriteLedger(ctx, task.JobId, stopChan); err != nil {
							clog.Error("write job ledger to blockchain error %v current job will terminate, please contact bountycloud@163.com", err)
							finishedChan <- 1
							endChan <- 1
							cancel()
							return
						}

						// 周期支付
						go circleBilling(ctx, task.JobId,stopChan)

						//TODO 执行真实的任务消费逻辑，从MessageBody imageId,JobId
						err = process(task, cpu, ram, stopChan)
						if err != nil {
							//TODO 发送这个任务处理失败消息，进行重试或者报告使用，需要发送json信息
							exceptionJob := &model.ExceptionJob{
								JobId:     task.JobId,
								Exception: err.Error(),
							}
							exceptionJobBytes, _ := json.Marshal(exceptionJob)
							// 目前异常暂时没有处理，目前异常要让用户手动处理
							err = RetryCommitByMq(string(exceptionJobBytes), Error)
							if err != nil {
								// 需要直接终止，目前上层捕获了，会终止程序，因为这一步出现问题很小，到时候手动处理
								panic(fmt.Errorf("bountyCloud received job exception msg error %v, please contact bountycloud@163.com", err))
							}
						}

						// 任务结束，终止计费
						cancel()
					}

					finishedChan <- 1
					endChan <- 1
				}
			case err := <-errChan:
				{
					// Topic中没有消息可消费。
					if strings.Contains(err.(errors.ErrCode).Error(), "MessageNotExist") {
						clog.Info("No new task, continue!")
					} else {
						clog.ListInfo([]string{"err"}, []interface{}{err})
						time.Sleep(time.Duration(5) * time.Second)
					}
					finishedChan <- 1
					endChan <- 1
				}
			case <-time.After(3600 * 24 * 365 * time.Second):
				// 单个任务的最长执行时间，可以让用户指定,目前先默认 24小时，这个可以防止一个人一直占用一个任务？
				{
					clog.Info("Timeout of consumer message ??")
					endChan <- 1
				}
			}
		}()

		// 拿消息，这个和上一个匿名函数不在一个里面
		go func(finishedChan chan int, respChan chan mq_http_sdk.ConsumeMessageResponse, errChan chan error, endChan chan int) {

			// 先消费一次
			mqConsumer.ConsumeMessage(respChan, errChan,
				1, // 一次最多消费1条，消费完才给下一个任务（最多可设置为16条）。
				3, // 长轮询时间3s（最多可设置为30s）。
			)

			// for也不需要，可以自己运行，这里本来就会阻塞住
			//for {
				select {
				case <- finishedChan:
					return
				//default:
				// 以前的用法有问题，不需要default
				//	time.Sleep(time.Second * 10)
				}
			//}
		}(finishedChan,respChan,errChan,endChan)

		// 阻塞，直到你设置的结束
		<-endChan
	}
}

// 只有Controller写Job成功，才可以进行下面的计费操作
// @Author: hyc
// @Description:
// @Date: 2022/1/29 11:14
func checkWriteLedger(ctx context.Context, jobId int, stopChan chan struct{}) error {

	for i := 0; i < 3; i++ {
		rep, err := client.RestyClient.R().
			SetQueryParam("jobId", strconv.Itoa(jobId)).
			SetCookies(constants.Cookies).
			Get(constants.JobBlockChainState)
		if err != nil {
			err =fmt.Errorf("can't connection to controller server")
			return err
		}

		// job blockchain 状态：0: 还未注册 1：注册成功 2: 注册失败(controller stopped)
		if rep.StatusCode() == 200 {
			jsonBody := utils.GetJsonBody(rep.Body())

			code := jsonBody["code"].(float64)
			clog.Info("current blockchain job register code is %v， please wait at most 10 minutes or terminate", code)
			if code == 1 {
				clog.Info("current job : %d has register blockchain job success, start to bill", jobId)
				return nil
			} else if code == 2 {
				err = fmt.Errorf("block chain job register error")
				return err
			} else {
				// 一直是 0 证明有问题的，已经要结束了 第三次，只重试三次
				if i == 2 {
					err = fmt.Errorf("block chain job register timeout, currnet job over")
					return err
				} else {
					time.Sleep(time.Second * time.Duration((i+1) * (i+1) * 30))
					continue
				}
			}
		} else {
			if i == 2 {
				err = fmt.Errorf("err, code: %v",rep.StatusCode())
				return err
			} else {
				time.Sleep(time.Second * 3)
				continue
			}
		}

	}

	return nil
}


func PushMsg(messageBody, messageTag string) error {

	// 循环发送4条消息。
	var msg mq_http_sdk.PublishMessageRequest

	msg = mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,         // 消息内容。
		MessageTag:  messageTag,          // 消息标签。
		Properties:  constants.JobProperties, // 消息属性，里面有uid用来唯一确认消息
	}
	// 设置消息的Key。
	msg.MessageKey = "MessageKey"
	// 设置消息自定义属性，目前看不需要设置属性
	//msg.Properties["a"] = strconv.Itoa(i)

	ret, err := mqProducer.PublishMessage(msg)

	if err != nil {
		clog.Error("err: %v", err)
		return err
	} else {
		clog.Debug("Publish ---->MessageId:%s, BodyMD5:%s, messageTag: %s", ret.MessageId, ret.MessageBodyMD5, messageTag)
	}
	//time.Sleep(time.Duration(100) * time.Millisecond)

	return nil
}

// 处理消息
// @Author: hyc
// @Description: msg格式为字典:  task_dict = {'task': task_image_name, 'task_developer':task_developer_id, 'job':job.id}
// @Date: 2021/11/16 12:45
func process(task *model.Task, cpu, ram int, stopChan chan struct{}) error {

	// 运行代码，开始执行任务
	taskResult, pullTime, runTime, err := run(task.ImagePath, cpu, ram, task.Ports, stopChan, task.JobId)
	if err != nil {
		clog.ListError([]string{"run task err"}, []interface{}{err})
		return err
	}

	// TODO taskResult处理，字符说超过5M大小，需要另赋值,目前先这样处理，避免日志不能传过去
	if len(taskResult) >= 4 * 1024 * 1024 {
		taskResult = constants.ResultTooLarge
	}

	resultSum, err := json.Marshal(model.Result{
		Result:    taskResult,
		PullTime:  pullTime,
		RunTime:   runTime,
		TotalTime: pullTime + runTime,
		JobId:     task.JobId,
	})
	if err != nil {
		clog.Error("err %v", err)
		return err
	}

	//TODO 将结果post给服务器，附带重试，需要带上管理员密码
	//re, err := retryCommitByHttp(resultSum)
	err = RetryCommitByMq(string(resultSum), Finished)
	if err != nil {
		clog.ListError([]string{"commit result fail err"}, []interface{}{err})
		return err
	}
	//clog.ListInfo([]string{"re"}, []interface{}{string(re.Body())})
	//TODO 我只负责队列消息，默认成功

	//fmt.Printf("%+v\n", msg)

	return nil
}

func circleBilling(ctx context.Context, jobId int,stopChan chan struct{}) {

	// 先进行一次计费防止不足十分钟这种，避免资金不足恶意调用
	stopChan, done := PrePayOne(jobId, stopChan)
	if done {
		return
	}

	insufficient_balance_count := 0

	// FIXME 这里需要改
	//timer := time.NewTicker(time.Minute * 10)
	timer := time.NewTicker(time.Second * 20)
	for {

		// 这样不行，只会循环一次
		select {
		case <-timer.C:
			// TODO 进行发送计费，setTime
			clog.Info("billing job id %v", strconv.FormatInt(int64(jobId), 10))
			re, err :=
				hfSetTimeAndPay(strconv.FormatInt(int64(jobId), 10), float64(10*60*1000), stopChan) // 我的计时单位是啥来着
			if err != nil {
				err = fmt.Errorf("billing err: %v, container will terminate right , if have any problem please send Email to bountycloud@163.com", err)
				clog.Error(err.Error())
				exceptionJob := &model.ExceptionJob{
					JobId:     jobId,
					Exception: err.Error(),
				}
				exceptionJobBytes, _ := json.Marshal(exceptionJob)
				RetryCommitByMq(string(exceptionJobBytes), Error)

				stopChan <- struct{}{}
				return
			}
			jsonResp := utils.GetJsonBody(re.Body())
			// FIXME 这里的处理可能有问题，目前没有这个错误信息了，所以不需要这样处理
			if !jsonResp["success"].(bool) {
				// 下面注释的代码目前看不需要，写的好乱呀，功能实现了再优化吧，现在没时间弄
				// FIXME 这个余额不足的逻辑需要小小的验证一下，不过应该没有啥问题
				if strings.Contains(jsonResp["message"].(string), "There is not enough money in account A") {
					// 余额不足的处理逻辑
					ackJob := &model.AckJob{
						JobId:    jobId,
						Username: constants.Username, // provider不重要
					}
					ackJobStr, _ := json.Marshal(ackJob)
					RetryCommitByMq(string(ackJobStr), InsufficientBalance)

					// 统计次数，如果是两次就停止return了
					insufficient_balance_count++
					if insufficient_balance_count == 3 {
						stopChan <- struct{}{}
						return
					} else {
						continue
					}
				}

				clog.Info("billing err: %v, container will terminate right , if have any problem please send Email to bountycloud@163.com", jsonResp)
				stopChan <- struct{}{}
				return
			}
		// 这个应该是发送多次导致的，有一个就能接收一个，就是只有cancel之后，才有done
		case <-ctx.Done():
			clog.Info("%+v, reason: %v\n", "The container runs at the end and billing ends", ctx.Err())
			return

			//default:
			// 不需要default，会自动阻塞重试
			//	//clog.Info("%+v\n", "The container is still running for the next round of billing")
			//	time.Sleep(time.Second * 10)
		}
	}

}

// PrePayOne 提前计费一，防止恶意调用
// @Author: hyc
// @Description:
// @Date: 2022/1/28 19:35
func PrePayOne(jobId int, stopChan chan struct{}) (chan struct{}, bool) {
	// TODO 进行发送计费，setTime
	clog.Info("PrePayOne: billing job id %v", strconv.FormatInt(int64(jobId), 10))
	re, err :=
		hfSetTimeAndPay(strconv.FormatInt(int64(jobId), 10), float64(10*60*1000), stopChan) // 我的计时单位是啥来着
	if err != nil {
		err = fmt.Errorf("PrePayOne billing err: %v, container will terminate right , if have any problem please send Email to bountycloud@163.com", err)
		clog.Error(err.Error())
		exceptionJob := &model.ExceptionJob{
			JobId:     jobId,
			Exception: err.Error(),
		}
		exceptionJobBytes, _ := json.Marshal(exceptionJob)
		RetryCommitByMq(string(exceptionJobBytes), Error)

		stopChan <- struct{}{}
		return nil, true
	}
	jsonResp := utils.GetJsonBody(re.Body())
	// FIXME 这里的处理可能有问题，目前没有这个错误信息了，所以不需要这样处理
	if !jsonResp["success"].(bool) {

		// FIXME 这个余额不足的逻辑需要小小的验证一下，不过应该没有啥问题
		if strings.Contains(jsonResp["message"].(string), "There is not enough money in account A") {
			// 余额不足的处理逻辑
			ackJob := &model.AckJob{
				JobId:    jobId,
				Username: constants.Username, // provider不重要
			}
			ackJobStr, _ := json.Marshal(ackJob)
			RetryCommitByMq(string(ackJobStr), InsufficientBalance)

			// 支付错误直接停止
			stopChan <- struct{}{}
			return nil, true
		}

		clog.Info("PrePayOne billing err: %v, container will terminate right , if have any problem please send Email to bountycloud@163.com", jsonResp)
		stopChan <- struct{}{}
		return nil, true
	}
	return stopChan, false
}

// hfSetTime 周期支付
// @Author: hyc
// @Description:
// @Date: 2021/12/6 21:00
func hfSetTimeAndPay(jobId string, time float64, stopChan chan struct{}) (*resty.Response, error) {

	validCost := math.Ceil(time * constants.TimeCoefficient / 100)

	var (
		re *resty.Response
		err error
	)

	// 先注册，因为过期是大概率事件，麻烦就麻烦点吧，计算时间用的是Provider这个用户名进行操作的，而且bc端有进行验证，获取新的token
	constants.GlobalToken, err = hyperledger.HfClient.RegisterUser(constants.Username, constants.UserOrg)
	if err != nil {
		clog.Error("register user error: %v", err)
		return nil, err
	}

	re, err = hyperledger.HfClient.InvokeTransaction(constants.GlobalToken, constants.ChannelName, constants.ChaincodeName, constants.UserOrg, "circleSetValidTimeCost", constants.AllOrg, []string{jobId, strconv.FormatFloat(validCost, 'f', 0, 64)})
	if err != nil {
		clog.Error("hf set circle time cost error: %v", err)
		return nil, err
	}

	//if strings.Contains(string(re.Body()), "jwt expired") ||
	//	strings.Contains(string(re.Body()), "jwt malformed") ||
	//	strings.Contains(string(re.Body()), "User was not found") ||
	//	strings.Contains(string(re.Body()), "UnauthorizedError") {
	//
	//
	//
	//	re, err = hyperledger.HfClient.InvokeTransaction(constants.GlobalToken, constants.ChannelName, constants.ChaincodeName, constants.UserOrg, "circleSetValidTimeCost", constants.AllOrg, []string{jobId, strconv.FormatFloat(validCost, 'f', 0, 64)})
	//
	//	if err != nil {
	//		clog.Error("hf set circle time cost error: %v", err)
	//		return nil, err
	//	}
	//}

	response := new(model.BlockResponse)
	err = json.Unmarshal(re.Body(),&response)
	if err != nil {
		return nil, err
	}

	if !response.Success && strings.Contains(response.Message,"The state for this job is already stop and can't' make money change") {
		// TODO 这里需要发送信号，进行结束动作，考虑用一个chan通知process进程主动结束，目前已经完成
		stopChan <- struct {}{}
		// 不需要panic，主动结束了，会有ctx的信号通知，结束计费
		//panic(response.Message)
	}


	return re, nil
}

// 重试提交 http 形式
// @Author: hyc
// @Description:
// @Date: 2021/11/16 19:29
func retryCommitByHttp(sum []byte) (*resty.Response, error) {

	for i := 0; i < 3; i++ {
		re, err := client.RestyClient.R().
			// FIXME 这里需要改成管理员密账号密码，不能用这个提交
			SetBasicAuth("forge2yc", "hou19941230l").
			SetBody(map[string]interface{}{
				"resultSum": sum,
			}).Post(constants.CommitResult)

		if err != nil {
			clog.ListError([]string{"err"}, []interface{}{err})
			time.Sleep(time.Duration(i) * time.Second)
			continue
		} else {
			return re, nil
		}
	}

	err := fmt.Errorf("commit result fail")
	return nil, err
}

// RetryCommitByMq 重试提交通过mq, 不处理返回结果，默认提交了就一定会成功
// @Author: hyc
// @Description: 目前只尝试4次，多了就负责了
// @Date: 2021/11/29 12:35
func RetryCommitByMq(body string, tag string) error {
	//clog.Info("current tag is %s", tag)
	for i := 0; i <= 3; i++ {

		err := PushMsg(body, tag)
		if err != nil {
			clog.Error("retry %d times err: %v", i, err)
			time.Sleep(time.Duration(i) * time.Second)
			continue
		} else {
			return nil
		}
	}

	err := fmt.Errorf("commit fail, tag: %v", tag)
	return err
}
