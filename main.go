package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/beck-8/subs-check/app"
)

func main() {
	application := app.New(fmt.Sprintf("%s-%s", Version, CurrentCommit))
	slog.Info(fmt.Sprintf("当前版本: %s-%s", Version, CurrentCommit))

	if err := application.Initialize(); err != nil {
		slog.Error(fmt.Sprintf("初始化失败: %v", err))
		os.Exit(1)
	}

	// --- 核心修改：异步运行并监控 ---
	
	// 1. 在后台协程中启动程序
	go func() {
		application.Run()
	}()

	slog.Info("程序已在后台启动，开始监控 output/base64.txt...")

	// 2. 主线程循环检查文件状态
	ticker := time.NewTicker(20 * time.Second)
	startTime := time.Now()
	
	for range ticker.C {
		// 检查是否超时（例如 50 分钟），防止死循环浪费 Actions 额度
		if time.Since(startTime) > 50*time.Minute {
			slog.Error("监控超时，强制退出")
			os.Exit(1)
		}

		// 检查目标文件
		info, err := os.Stat("output/base64.txt")
		if err == nil {
			// 如果文件存在且大小大于 1KB，认为筛选已初步完成
			if info.Size() > 1024 {
				// 给程序一点点时间完成最后的 IO 写入
				time.Sleep(5 * time.Second)
				slog.Info(fmt.Sprintf("检测到结果文件生成 (大小: %d 字节)，准备退出...", info.Size()))
				os.Exit(0) // 优雅退出，触发 GitHub Actions 的下一步
			}
		}
		
		slog.Info("正在筛选节点中，请稍候...")
	}
}
