package util

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"
)

// portScanner 端口扫描仪
/* portScanner尝试与指定目标的端口建立TCP连接。
如果在1秒内成功建立连接，则表明端口开放，函数返回true。
如果连接失败或超时，则表明端口关闭或不可达，函数返回false。

参数:
  target string - 目标主机或IP地址。
  port int - 要扫描的端口号。

返回值:
  bool - 如果端口开放则返回true，否则返回false。
*/
func portScanner(target string, port int) bool {
	// 尝试与目标端口建立TCP连接，设置1秒的超时时间。
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", target, port), 1*time.Second)
	if err != nil {
		// 连接失败，端口可能关闭或不可达。
		return false
	}
	// 关闭连接，因为我们只是进行一次简单的可达性检查。
	conn.Close()
	// 成功建立连接，端口开放。
	return true
}

// getAllIPs 根据输出的网段计算出该网段的所有ip
/*
getAllIPs 函数根据给定的 CIDR 编址范围获取所有 IP 地址，并以字符串切片的形式返回。
参数:
  cidr string - 表示 CIDR 编址范围的字符串，例如 "192.168.1.0/24"。
返回值:
  []string - 在给定 CIDR 编址范围内的所有 IP 地址的字符串切片。
*/
func getAllIPs(cidr string) []string {
	var ips []string
	// 解析 CIDR 字符串，获取 IP 网络对象
	_, ipv4Net, err := net.ParseCIDR(cidr)
	if err != nil {
		// 如果解析出错，则打印错误信息并终止程序
		log.Panicln(err)
	}
	// 将 IPNet 结构体的掩码和地址转换为 uint32 类型
	// 由于网络字节序为 BigEndian，因此使用 BigEndian
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)

	// 计算最后一个地址
	// 通过将起始地址与掩码进行与操作，然后将结果与 (掩码 ^ 0xffffffff) 进行或操作来得到最后一个地址
	finish := (start & mask) | (mask ^ 0xffffffff)

	// 遍历地址范围内的所有地址
	for i := start; i <= finish; i++ {
		// 将 uint32 类型的地址转换回 net.IP 类型
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		// 将 IP 地址字符串添加到结果切片中
		ips = append(ips, ip.String())
	}
	return ips
}

// pingHostScanner
/*
 pingHostScanner 函数用于通过执行ping命令来检查指定目标是否可达。
参数 target 表示需要进行ping检测的目标主机地址。
返回值表示是否成功ping通了目标主机，true表示可达，false表示不可达。
*/
func pingHostScanner(target string) bool {
	// 使用context设置1秒的超时时间，确保ping操作不会无限进行。
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	// 确保在函数退出时取消context，避免资源泄露。
	defer cancel()
	// 构建ping命令，发送1个ICMP报文到目标主机。
	cmd := exec.CommandContext(ctx, "ping", []string{"-c", "1", target}...)
	// 使用bytes.Buffer来存储cmd的stdout和stderr输出，这里我们并不关心输出内容。
	// var stdout, stderr bytes.Buffer
	// cmd.Stdout = &stdout
	// cmd.Stderr = &stderr
	// 执行ping命令。
	err := cmd.Run()
	// 检查执行过程中是否超时或出现其他错误。
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || err != nil {
		// 超时或错误则返回false。
		return false
	}
	// 执行成功（无超时且无错误）则返回true。
	return true
}
