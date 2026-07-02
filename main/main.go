package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-ping/ping"
)

type Pinger struct {
	mu      sync.RWMutex
	targets map[string]context.CancelFunc // IP -> cancel func
	wg      sync.WaitGroup
}

func NewPinger() *Pinger {
	return &Pinger{
		targets: make(map[string]context.CancelFunc),
	}
}

// 添加要 Ping 的 IP
func (p *Pinger) AddTarget(ip string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.targets[ip]; exists {
		fmt.Printf("IP %s 已在监控中\n", ip)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.targets[ip] = cancel
	p.wg.Add(1)

	go p.pingLoop(ctx, ip)
	fmt.Printf("已启动对 %s 的监控 (每60秒一次)\n", ip)
}

// 删除 IP
func (p *Pinger) RemoveTarget(ip string) {
	p.mu.Lock()
	cancel, exists := p.targets[ip]
	if exists {
		cancel()
		delete(p.targets, ip)
	}
	p.mu.Unlock()

	if exists {
		fmt.Printf("已停止对 %s 的监控\n", ip)
	}
}

// 单个 IP 的 Ping 循环
func (p *Pinger) pingLoop(ctx context.Context, ip string) {
	defer p.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	fmt.Printf("[%s] 开始监控\n", ip)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] 监控已停止\n", ip)
			return
		case <-ticker.C:
			p.doPing(ip)
		}
	}
}

// 使用 go-ping 库执行 Ping
func (p *Pinger) doPing(ip string) {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		fmt.Printf("[%s] %s ❌ 创建Pinger失败: %v\n", time.Now().Format("15:04:05"), ip, err)
		return
	}

	pinger.Count = 1 // 只发1个包
	pinger.Timeout = 3 * time.Second
	pinger.SetPrivileged(true) // Windows 下建议设为 true（需要管理员权限）

	err = pinger.Run()
	if err != nil {
		fmt.Printf("[%s] %s ❌ Ping失败: %v\n", time.Now().Format("15:04:05"), ip, err)
		return
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv > 0 {
		fmt.Printf("[%s] %s ✅ Ping成功 RTT: %v\n",
			time.Now().Format("15:04:05"), ip, stats.AvgRtt)
	} else {
		fmt.Printf("[%s] %s ❌ Ping失败 (0 packets received)\n",
			time.Now().Format("15:04:05"), ip)
	}
}

func main() {
	pinger := NewPinger()

	// 示例：添加几个 IP
	pinger.AddTarget("172.19.0.21")
	pinger.AddTarget("8.8.8.8")
	pinger.AddTarget("1.1.1.1")

	// 命令行交互
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("输入 IP 添加监控，输入 rm:IP 删除，输入 q 退出")
		for scanner.Scan() {
			input := scanner.Text()
			if input == "q" || input == "quit" {
				break
			}
			if len(input) > 3 && input[:3] == "rm:" {
				ip := input[3:]
				pinger.RemoveTarget(ip)
			} else if input != "" {
				pinger.AddTarget(input)
			}
		}
	}()

	// 优雅退出
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("\n正在停止所有监控...")
	pinger.mu.Lock()
	for ip, cancel := range pinger.targets {
		cancel()
		fmt.Printf("停止 %s\n", ip)
	}
	pinger.mu.Unlock()

	pinger.wg.Wait()
	fmt.Println("程序已退出")
}
