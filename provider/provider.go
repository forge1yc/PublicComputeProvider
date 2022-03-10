package provider

import (
	"ComputeProviderByGo/model"
	"encoding/json"
	"fmt"
	"github.com/forge1yc/clog/clog"
	"time"
)

// RunDocker 运行docker
// @Author: hyc
// @Description:
// @Date: 2021/11/13 23:07
func RunDocker(imagePath string) {

	//resp, err := cli.ContainerCreate(ctx, &container.Config{
	//	Image: imageName,
	//	Cmd:   []string{"echo", "hello world"},
	//}, nil, nil, "",fmt.Sprintf("%s_%s","provider_container",time.Now().Unix()))
	//if err != nil {
	//	panic(err)
	//}
	//
	//
	//containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//
	//for _, container := range containers {
	//	clog.Info("%s %s\n", container.ID[:10], container.Image)
	//}
	//image :=



}

// DeleteContainerAndImage 删除镜像和容器
// @Author: hyc
// @Description:
// @Date: 2021/11/13 23:08
func DeleteContainerAndImage() {

}

// HFSetTime 区块链设置时间
// @Author: hyc
// @Description:
// @Date: 2021/11/13 23:09
func HFSetTime() {

}

// OnRequest 处理请求
// @Author: hyc
// @Description:
// @Date: 2021/11/13 23:10
func OnRequest() {

}

// Start 开始服务
// @Author: hyc
// @Description:
// @Date: 2021/11/14 12:16
func Start(user model.User) {
	defer func() {
		if err := recover(); err != nil {
			// 异常导出
			clog.Error("err: %v ",err)
		}
	}()

	//TODO 这里需要启动一个待cookie的会话，能够保持一直在线那种，目前看不需要那种了，我都用mq代替了
	go func() {
		err := heartBeatCheck(user)
		if err != nil {
			err = fmt.Errorf("heartBeat check err: %v",err)
			// 上层会捕获
			panic(err)
		}
	}()


}

// HeartBeatCheck provider 的心跳检测
// @Author: hyc
// @Description: 60分钟一次的心跳检测
// @Date: 2021/11/29 15:11
func heartBeatCheck(user model.User) error {
	u, err := json.Marshal(user)
	if err != nil {
		clog.Info("marshal user error: %v",err)
		return err
	}
	for {
		err = RetryCommitByMq(string(u),State)
		if err != nil {
			clog.Error("err: %v",err)
		}
		// 10分钟一次探活
		time.Sleep(10 * time.Minute)
		//time.Sleep(1 * time.Second)
	}
}


func ConfirmReceived(startTime int64,username string) error {

	//RetryCommitByMq()
	return nil
}


























