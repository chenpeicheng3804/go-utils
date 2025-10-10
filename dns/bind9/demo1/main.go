// main.go
package demo1

import (
	"fmt"
	"log"
)

func test1() {
	// 创建DNS客户端
	client := NewDNSClient(
		"10.10.14.20", // DNS服务器地址
		"dnsadmin",    // TSIG密钥名称
		"pXENbHSMraRrVnnVkSPIgpmZLj/kJNu/lnHUkTFkKMw=", // TSIG密钥
		"53", // 端口
	)

	// 示例：添加记录
	fmt.Println("=== 添加记录 ===")
	err := client.AddRecord(
		"example.com",
		"test.example.com",
		"A",
		"192.168.1.100",
		300,
	)
	if err != nil {
		log.Printf("添加记录失败: %v", err)
	} else {
		fmt.Println("添加记录成功")
	}

	// 示例：修改记录
	fmt.Println("\n=== 修改记录 ===")
	err = client.UpdateRecord(
		"example.com",
		"test.example.com",
		"A",
		"192.168.1.200",
		300,
	)
	if err != nil {
		log.Printf("修改记录失败: %v", err)
	} else {
		fmt.Println("修改记录成功")
	}

	// 示例：删除记录
	fmt.Println("\n=== 删除记录 ===")
	err = client.DeleteRecord(
		"example.com",
		"test.example.com",
		"A",
		300,
	)
	if err != nil {
		log.Printf("删除记录失败: %v", err)
	} else {
		fmt.Println("删除记录成功")
	}

	// 示例：同步记录
	fmt.Println("\n=== 同步记录 ===")
	records := []DNSRecord{
		{Name: "www.example.com", Type: "A", Value: "192.168.1.10", TTL: 300},
		{Name: "mail.example.com", Type: "A", Value: "192.168.1.20", TTL: 300},
		{Name: "example.com", Type: "MX", Value: "10 mail.example.com.", TTL: 300},
	}

	err = client.SyncRecords("example.com", records)
	if err != nil {
		log.Printf("同步记录失败: %v", err)
	} else {
		fmt.Println("同步记录成功")
	}
}
func test2() {
	// 创建基础DNS客户端
	client := NewDNSClient(
		"10.10.14.20", // DNS服务器地址
		"dnsadmin",    // TSIG密钥名称
		"pXENbHSMraRrVnnVkSPIgpmZLj/kJNu/lnHUkTFkKMw=", // TSIG密钥
		"53", // 端口
	)

	// 创建支持同步的客户端
	syncClient := NewSyncDNSClient(client)

	// 定义本地记录
	localRecords := []DNSRecord{
		{Name: "www.example.com", Type: "A", Value: "192.168.1.10", TTL: 300},
		{Name: "mail.example.com", Type: "A", Value: "192.168.1.20", TTL: 300},
		{Name: "ftp.example.com", Type: "CNAME", Value: "www.example.com", TTL: 300},
		{Name: "example.com", Type: "MX", Value: "10 mail.example.com", TTL: 300},
	}

	// 方式1：从服务器同步到本地（服务器端同步到本地）
	fmt.Println("=== 从服务器同步到本地 ===")
	remoteRecords, err := syncClient.SyncFromServer("example.com")
	if err != nil {
		log.Printf("从服务器同步失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 条记录\n", len(remoteRecords))
	}

	// 方式2：同步本地数据到服务器（同步本地数据到服务器）
	fmt.Println("\n=== 同步本地数据到服务器 ===")
	err = syncClient.SyncToServer("example.com", localRecords)
	if err != nil {
		log.Printf("同步到服务器失败: %v", err)
	} else {
		fmt.Println("同步完成")
	}

	// 方式3：双向同步
	fmt.Println("\n=== 双向同步 ===")
	_, err = syncClient.BidirectionalSync("example.com", localRecords)
	if err != nil {
		log.Printf("双向同步失败: %v", err)
	}

	// 方式4：比较差异
	fmt.Println("\n=== 比较差异 ===")
	if len(remoteRecords) > 0 {
		diff := syncClient.CompareRecords(localRecords, remoteRecords)
		diff.PrintDiff()
	}
}
