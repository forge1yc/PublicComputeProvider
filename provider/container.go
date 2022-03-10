package provider

import (
	"ComputeProviderByGo/constants"
	"ComputeProviderByGo/model"
	"ComputeProviderByGo/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/forge1yc/clog/clog"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	GB            = 1024 * 1024 * 1024
	BanMemorySwap = -1
	VCpu          = 100000
)



func run(imageId string, cpu, ram int, ports []string, stopChan chan struct{}, jobId int) (string, int64, int64, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		clog.ListError([]string{"create docker client err"}, []interface{}{err})
		return "", 0, 0, err
	}
	// 这个login可以重复登录，用来私人仓库的使用，主要是为了安全作用，这里，目前看可以直接登录使用，会保存在用户本地将token，目前看有比没有强，问题不大
	//re,err := cli.RegistryLogin(context.Background(),types.AuthConfig{
	//	Username:      "username",
	//	//Password:      "",
	//	Auth:          "token",
	//})
	//if re.IdentityToken


	//cli, err := client.NewClient("tcp://192.168.100.33:2376", "v1.12", nil, nil)
	defer func(cli *client.Client) {
		err := cli.Close()
		if err != nil {
			clog.ListError([]string{"close client error:"}, []interface{}{err})
		}
	}(cli)

	id, pullTime, err := createContainer(cli, imageId, cpu, ram, ports)
	if err != nil {
		clog.ListError([]string{"err"}, []interface{}{err})
		//这里终止前需要提示用户，job异常结束
		exceptionJob := &model.ExceptionJob{
			JobId:     jobId,
			Exception: err.Error(),
		}
		exceptionJobBytes, _ := json.Marshal(exceptionJob)
		// 目前异常暂时没有处理，目前异常要让用户手动处理
		err = RetryCommitByMq(string(exceptionJobBytes), Error)
		if err != nil {
			// 需要直接终止，目前上层捕获了，会终止程序，因为这一步出现问题很小，到时候手动处理
			panic(fmt.Errorf("bountyCloud received job exception msg error %v, please contact bountycloud@163.com", err))
		}
		log(err)
		return "", 0, 0, err
	}

	// FIXME 退出的时候最后要清理
	defer func(id string, cli *client.Client, imageId string) {
		//stopContainer(id,cli)
		//removeImages(cli,imageId)
		// 删除容器(这里不需要主动停止，运行完已经是终止状态了）
		id, err = removeContainer(id, cli)
		if err != nil {
			clog.Error("删除容器 %v 失败，请手动删除", id)
			//fmt.Println("删除容器", id, "成功")
		} else {
			clog.Info("删除容器: %v 成功", id)
		}

		// FIXME 删除镜像, 最后需要补上， 这个删除镜像有问题，目前先不打开，非debug情况下，才删除
		if !constants.DEBUG {
			err = removeImages(cli, imageId)
			if err != nil {
				clog.Error("删除镜像 %v 失败，请手动删除", imageId)
				//fmt.Println("删除容器", id, "成功")
			} else {
				clog.Info("删除镜像: %v 成功", imageId)
			}
		}

	}(id,cli,imageId)

	// 必须先开始
	runTimeStart := startContainer(id, cli)

	// 一个匿名函数，如果发现了有终止信号，需要及时进行任务的终止，终止信号由Controller发出，在启动之前调用，会让wait不在等待，往下运行|| 找到原因了，不能未开始就stop
	go func(id string, cli *client.Client, stopChan chan struct{}) {
		// 及时关闭 FIXME 这里有问题，被关闭了，竟然还会发送，不可以的，这样，会引起panic
		//defer close(stopChan)
		// 这里能够调用到结束，但是实际还在运行，咋回事，我手动结束之后才能记录log日志，进行下一个任务接收工作
		if _, ok := <-stopChan; ok {
			stopContainer(id,cli)
		}

	}(id, cli, stopChan)

	// 等待容器结束
	_ = ContainerWait(context.Background(), cli, id)

	runTime := time.Since(runTimeStart).Milliseconds()
	out, err := cli.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
	})
	if err != nil {
		clog.ListError([]string{"err"}, []interface{}{err})
		return "", 0, 0, err
	}

	successBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)
	_, _ = stdcopy.StdCopy(successBuffer, errorBuffer, out)
	//fmt.Printf("successsBuffer: %+v\n",successBuffer.String())
	//fmt.Printf("errorBuffer: %+v\n",errorBuffer.String())

	clog.Info("当前任务运行结束，等待下一个")
	return utils.MergeStdOutAndErrOut(successBuffer, errorBuffer), pullTime, runTime, nil
}

// ContainerWait wait 等待结束
// stop 是 exited 状态
// @Author: hyc
// @Description:
// @Date: 2021/11/15 21:07
func ContainerWait(ctx context.Context, cli *client.Client, containerId string) error {
	statusCh, errCh := cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
	return nil
}

// 列出镜像
func listImage(cli *client.Client) {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	log(err)

	for _, image := range images {
		fmt.Println(image)
	}
}

// FIXME 这部分最后需要补充到Provider的源码里面，我的Provider才应该是大书特书的地方!!!!!!!，我的Provider除了代码不完美，逻辑很是完美
// 创建容器
func createContainer(cli *client.Client, imagePath string, cpu, ram int, ports []string) (string, int64, error) {
	start := time.Now()
	ctx := context.Background()
	//cli.ImagePull(ctx, imagePath, types.ImagePullOptions{})

	reader, err := cli.ImagePull(ctx, imagePath, types.ImagePullOptions{})
	if err != nil {
		clog.Error("pull image %s fail, err, please check imagePath", imagePath)
		return "", 0, err
	}

	// 记录容器日志
	io.Copy(os.Stdout, reader)

	// 构建端口映射配置
	config, portMap := setBindPort(ports, imagePath)

	// 指定宿主机内存和CPU的使用以及端口映射
	hostConfig := &container.HostConfig{
		PortBindings: portMap,
		Resources: container.Resources{
			Memory:     int64(ram) * GB,
			CPUPeriod:  100000,
			CPUQuota:   int64(cpu) * VCpu,
			MemorySwap: BanMemorySwap,
		},
	}

	containerName := utils.LastWord(utils.ChangeImagePathToName(imagePath))
	body, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, containerName)
	log(err)
	fmt.Printf("Container ID: %s\n", body.ID)

	pullTime := time.Since(start)

	return body.ID, pullTime.Milliseconds(), nil
}

// 启动
func startContainer(containerID string, cli *client.Client) time.Time {
	start := time.Now()

	err := cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	log(err)
	if err == nil {
		fmt.Println("容器", containerID, "启动成功，等待计费")
	}

	return start
}

// 停止容器，当检测到stop状态时，在chan中接收到信号后，就主动停止了，简单，明天弄
func stopContainer(containerID string, cli *client.Client) {
	timeout := time.Second * 10
	err := cli.ContainerStop(context.Background(), containerID, &timeout)
	if err != nil {
		log(err)
	} else {
		fmt.Printf("容器%s已经被停止\n", containerID)
	}
}

// 删除容器
func removeContainer(containerID string, cli *client.Client) (string, error) {
	err := cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	log(err)
	return containerID, err
}

// 删除镜像
func removeImages(cli *client.Client, imageId string) error {
	re, err := cli.ImageRemove(context.Background(), imageId, types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil {
		clog.Info("remove image fail, err %v", err)
		//fmt.Printf("remove image fail, err %v", err)
		return err
	}
	clog.Info("已删除镜像%s,确认结果: %+v\n", imageId, re)
	//fmt.Printf("已删除镜像%s,确认结果: %+v\n", imageId, re)
	return nil
}

// 打印日志

func log(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}
}

func convertPPortsToSlice(pport string) ([]string, error) {
	p := strings.Split(pport, "/")
	if len(p) < 2 {
		return nil, fmt.Errorf("format error")
	}
	return p, nil
}

// 设置宿主机与容器的端口映射,这个配置可能有问题
// @Author: hyc
// @Description: 如果传过来的是0，或者不规范，就配置 443/tcp 这样不会出现大的问题，其实可以选择置空
// @Date: 2021/12/13 16:54
func setBindPort(ports []string, imagePath string) (*container.Config, nat.PortMap) {

	var (
		exports = make(nat.PortSet, 10)
		portMap = make(nat.PortMap, 0)
		err error
	)
	for _, pport := range ports {
		portProtocol := make([]string,0)
		if pport == "0" {
			portProtocol = []string{"443", "tcp"}
		} else {
			portProtocol, err = convertPPortsToSlice(pport)
			if err != nil {
				clog.Error("protocol and port bind err:%v, set to default 443/tcp", err)
				// 有问题的时候设置443
				portProtocol = []string{"443", "tcp"}
			}
		}

		port, err := nat.NewPort(portProtocol[1], portProtocol[0])
		if err != nil {
			clog.Error("err:%v", err)
		}

		//exports = make(nat.PortSet, 10)
		exports[port] = struct{}{}

		//如果用户是公网用户，这里可以考虑对外提供服务(间歇性的提供服务，按天计数）
		portBind := nat.PortBinding{HostPort: portProtocol[0]}
		//portMap := make(nat.PortMap, 0)
		// 这里我的理解是一个容器的端口，可以对应宿主机的多个端口，也就是宿主机不同的端口都能够映射进来
		tmp := make([]nat.PortBinding, 0, 1)
		tmp = append(tmp, portBind)
		portMap[port] = tmp
	}
	config := &container.Config{
		Image:        imagePath,
		ExposedPorts: exports,
	}
	return config, portMap

}
