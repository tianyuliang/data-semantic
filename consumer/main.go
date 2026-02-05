// Package main Kafka消费者入口
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/kweaver-ai/dsg/services/apps/data-semantic/consumer/internal/data_understanding"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/consumer/internal/config"
	"github.com/kweaver-ai/dsg/services/apps/data-semantic/consumer/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/consumer.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化日志
	logx.MustSetup(logx.LogConf{
		ServiceName: c.Log.ServiceName,
		Mode:        c.Log.Mode,
		Level:       c.Log.Level,
	})

	logx.Infof("启动 %s 服务...", c.Name)

	// 初始化服务上下文
	svcCtx := svc.NewServiceContext(c)

	// 创建Kafka消费者
	brokers := c.MQ.Kafka.Brokers
	groupID := c.MQ.Kafka.Group
	topic := c.MQ.Kafka.Topic

	consumer, err := data_understanding.NewKafkaConsumer(brokers, groupID, []string{topic})
	if err != nil {
		logx.Errorf("创建Kafka消费者失败: %v", err)
		os.Exit(1)
	}

	// 注册消息处理器
	handler := data_understanding.NewDataUnderstandingHandler(svcCtx)
	consumer.RegisterHandler(topic, handler)

	// 创建context用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动消费者（在goroutine中）
	go func() {
		if err := consumer.Start(ctx); err != nil {
			logx.Errorf("Kafka消费失败: %v", err)
		}
	}()

	logx.Infof("Kafka消费者已启动，订阅主题: %s, 消费者组: %s", topic, groupID)

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logx.Info("收到退出信号，正在关闭...")
	cancel()
	logx.Info("服务已停止")
}
