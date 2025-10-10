// sync.go
package demo1

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// SyncDNSClient 支持同步功能的扩展客户端
type SyncDNSClient struct {
	*DNSClient
}

// NewSyncDNSClient 创建支持同步的DNS客户端
func NewSyncDNSClient(client *DNSClient) *SyncDNSClient {
	return &SyncDNSClient{
		DNSClient: client,
	}
}

// FetchRecords 从服务器获取区域的所有记录（服务器端同步到本地）
func (c *SyncDNSClient) FetchRecords(zone string) ([]DNSRecord, error) {
	zoneFQDN := dns.Fqdn(zone)

	// 创建区域传输客户端
	transfer := &dns.Transfer{}

	// 创建AXFR查询消息
	axfrMsg := &dns.Msg{}
	axfrMsg.SetAxfr(zoneFQDN)
	axfrMsg.SetTsig(c.tsigKeyName, dns.HmacSHA256, 300, time.Now().Unix())

	// 设置TSIG密钥
	transfer.TsigSecret = map[string]string{c.tsigKeyName: c.Secret}

	serverAddr := fmt.Sprintf("%s:%s", c.Server, c.Port)

	// 执行区域传输
	env, err := transfer.In(axfrMsg, serverAddr)
	if err != nil {
		return nil, fmt.Errorf("AXFR传输失败: %v", err)
	}

	var records []DNSRecord
	// 处理传输返回的记录
	for envelope := range env {
		if envelope.Error != nil {
			return nil, fmt.Errorf("AXFR传输错误: %v", envelope.Error)
		}

		for _, rr := range envelope.RR {
			// 跳过SOA记录（通常由服务器管理）
			if _, ok := rr.(*dns.SOA); ok {
				continue
			}

			// 解析记录信息
			record := c.parseRR(rr)
			if record != nil {
				records = append(records, *record)
			}
		}
	}

	return records, nil
}

// parseRR 解析DNS资源记录为DNSRecord结构
func (c *SyncDNSClient) parseRR(rr dns.RR) *DNSRecord {
	header := rr.Header()

	// 移除末尾的点以保持一致性
	name := strings.TrimSuffix(header.Name, ".")

	record := &DNSRecord{
		Name: name,
		TTL:  header.Ttl,
	}

	switch r := rr.(type) {
	case *dns.A:
		record.Type = "A"
		record.Value = r.A.String()
	case *dns.AAAA:
		record.Type = "AAAA"
		record.Value = r.AAAA.String()
	case *dns.CNAME:
		record.Type = "CNAME"
		record.Value = strings.TrimSuffix(r.Target, ".")
	case *dns.MX:
		record.Type = "MX"
		record.Value = fmt.Sprintf("%d %s", r.Preference, strings.TrimSuffix(r.Mx, "."))
	case *dns.TXT:
		record.Type = "TXT"
		// 合并TXT记录的所有字符串部分
		var txtData strings.Builder
		for i, txt := range r.Txt {
			if i > 0 {
				txtData.WriteString(" ")
			}
			txtData.WriteString(txt)
		}
		record.Value = txtData.String()
	case *dns.NS:
		record.Type = "NS"
		record.Value = strings.TrimSuffix(r.Ns, ".")
	case *dns.PTR:
		record.Type = "PTR"
		record.Value = strings.TrimSuffix(r.Ptr, ".")
	default:
		// 对于其他记录类型，使用默认处理方式
		rrStr := rr.String()
		parts := strings.Fields(rrStr)
		if len(parts) >= 4 {
			record.Type = parts[2]
			record.Value = strings.Join(parts[3:], " ")
		} else {
			// 如果无法解析，返回nil跳过该记录
			return nil
		}
	}

	return record
}

// SyncFromServer 从服务器同步到本地（服务器端同步到本地）
func (c *SyncDNSClient) SyncFromServer(zone string) ([]DNSRecord, error) {
	fmt.Printf("正在从服务器 %s 同步区域 %s 的记录...\n", c.Server, zone)

	records, err := c.FetchRecords(zone)
	if err != nil {
		return nil, fmt.Errorf("从服务器获取记录失败: %v", err)
	}

	fmt.Printf("成功同步 %d 条记录\n", len(records))
	return records, nil
}

// SyncToServer 将本地数据同步到服务器（同步本地数据到服务器）
func (c *SyncDNSClient) SyncToServer(zone string, records []DNSRecord) error {
	fmt.Printf("正在同步 %d 条记录到服务器 %s 的区域 %s ...\n", len(records), c.Server, zone)

	err := c.SyncRecords(zone, records)
	if err != nil {
		return fmt.Errorf("同步记录到服务器失败: %v", err)
	}

	fmt.Println("记录同步成功")
	return nil
}

// CompareRecords 比较两组记录的差异
func (c *SyncDNSClient) CompareRecords(local, remote []DNSRecord) *SyncDiff {
	diff := &SyncDiff{
		Add:    []DNSRecord{},
		Delete: []DNSRecord{},
		Update: []DNSRecord{},
	}

	// 将记录转换为map便于比较
	localMap := make(map[string]DNSRecord)
	remoteMap := make(map[string]DNSRecord)

	// 构建本地记录映射
	for _, record := range local {
		key := fmt.Sprintf("%s|%s|%s", record.Name, record.Type, record.Value)
		localMap[key] = record
	}

	// 构建远程记录映射
	for _, record := range remote {
		key := fmt.Sprintf("%s|%s|%s", record.Name, record.Type, record.Value)
		remoteMap[key] = record
	}

	// 找出需要添加的记录（在本地但不在远程）
	for key, record := range localMap {
		if _, exists := remoteMap[key]; !exists {
			diff.Add = append(diff.Add, record)
		}
	}

	// 找出需要删除的记录（在远程但不在本地）
	for key, record := range remoteMap {
		if _, exists := localMap[key]; !exists {
			diff.Delete = append(diff.Delete, record)
		}
	}

	return diff
}

// SyncDiff 同步差异结构
type SyncDiff struct {
	Add    []DNSRecord // 需要添加的记录
	Delete []DNSRecord // 需要删除的记录
	Update []DNSRecord // 需要更新的记录
}

// PrintDiff 打印同步差异
func (d *SyncDiff) PrintDiff() {
	if len(d.Add) > 0 {
		fmt.Println("需要添加的记录:")
		for _, record := range d.Add {
			fmt.Printf("  + %s %s %s (TTL: %d)\n", record.Name, record.Type, record.Value, record.TTL)
		}
	}

	if len(d.Delete) > 0 {
		fmt.Println("需要删除的记录:")
		for _, record := range d.Delete {
			fmt.Printf("  - %s %s %s (TTL: %d)\n", record.Name, record.Type, record.Value, record.TTL)
		}
	}

	if len(d.Update) > 0 {
		fmt.Println("需要更新的记录:")
		for _, record := range d.Update {
			fmt.Printf("  ~ %s %s %s (TTL: %d)\n", record.Name, record.Type, record.Value, record.TTL)
		}
	}

	if len(d.Add) == 0 && len(d.Delete) == 0 && len(d.Update) == 0 {
		fmt.Println("没有发现差异，记录已同步")
	}
}

// BidirectionalSync 双向同步（先从服务器同步到本地，再同步本地到服务器）
func (c *SyncDNSClient) BidirectionalSync(zone string, localRecords []DNSRecord) ([]DNSRecord, error) {
	fmt.Printf("开始双向同步区域 %s ...\n", zone)

	// 1. 从服务器获取当前记录
	remoteRecords, err := c.FetchRecords(zone)
	if err != nil {
		return nil, fmt.Errorf("获取服务器记录失败: %v", err)
	}

	// 2. 比较差异
	diff := c.CompareRecords(localRecords, remoteRecords)
	diff.PrintDiff()

	// 3. 同步本地记录到服务器
	err = c.SyncToServer(zone, localRecords)
	if err != nil {
		return nil, fmt.Errorf("同步到服务器失败: %v", err)
	}

	fmt.Println("双向同步完成")
	return remoteRecords, nil
}
