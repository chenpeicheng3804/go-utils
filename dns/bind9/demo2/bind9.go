package demo2

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSClient DNS客户端
type DNSClient struct {
	Server      string
	KeyName     string
	Secret      string
	Port        string
	client      *dns.Client
	tsigKeyName string
}
type DNSRecordInfo struct {
	Domain   string
	Name     string
	Type     string
	NewValue string
	OldValue string
	TTL      uint32
}

// DNSRecord DNS记录结构
type DNSRecord struct {
	Name  string
	Type  string
	Value string
	TTL   uint32
}

// SyncDiff 同步差异结构
type SyncDiff struct {
	Add    []DNSRecord // 需要添加的记录
	Delete []DNSRecord // 需要删除的记录
	Update []DNSRecord // 需要更新的记录
}

// NewDNSClient 创建新的DNS客户端
func NewDNSClient(server, keyName, secret, port string) *DNSClient {
	if port == "" {
		port = "53"
	}

	// 确保密钥名称是FQDN格式
	tsigKeyName := dns.Fqdn(keyName)

	client := &dns.Client{Net: "tcp"}
	client.TsigSecret = map[string]string{tsigKeyName: secret}

	return &DNSClient{
		Server:      server,
		KeyName:     keyName,
		Secret:      secret,
		Port:        port,
		client:      client,
		tsigKeyName: tsigKeyName,
	}
}

// dns string to type
func dnsStringToType(rtype string) uint16 {
	switch rtype {
	case "A":
		// A - IPv4地址记录
		return dns.TypeA
	case "AAAA":
		// AAAA - IPv6地址记录
		return dns.TypeAAAA
	case "CNAME":
		// CNAME - 规范名称记录，用于域名别名
		return dns.TypeCNAME
	case "MX":
		// MX - 邮件交换记录，指定邮件服务器
		return dns.TypeMX
	case "NS":
		// NS - 名称服务器记录，指定域名的DNS服务器
		return dns.TypeNS
	case "PTR":
		// PTR - 指针记录，用于反向DNS查找
		return dns.TypePTR
	case "TXT":
		// TXT - 文本记录
		return dns.TypeTXT
	case "SRV":
		// SRV - 服务定位记录
		return dns.TypeSRV
	case "CAA":
		// CAA - 证书颁发机构授权记录
		return dns.TypeCAA
	case "DNSKEY":
		// DNSKEY - DNS密钥记录
		return dns.TypeDNSKEY
	case "DS":
		// DS - 委派签名记录
		return dns.TypeDS
	case "NAPTR":
		// NAPTR - 命名权威指针记录
		return dns.TypeNAPTR
	case "SSHFP":
		// SSHFP - SSH公钥指纹记录
		return dns.TypeSSHFP
	case "TLSA":
		// TLSA - TLSA证书关联记录
		return dns.TypeTLSA
	default:
		return dns.TypeNone
	}
}

// 查询解析记录
func (d *DNSClient) QueryRecord(name string, rtype string) ([]dns.RR, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), dnsStringToType(rtype))
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, d.Server+":53")
	if err != nil {
		return nil, err
	}

	return r.Answer, nil
}

// createUpdateMessage 创建DNS更新消息
func (c *DNSClient) createUpdateMessage(zone string) *dns.Msg {
	zoneFQDN := dns.Fqdn(zone)
	msg := new(dns.Msg)
	msg.SetUpdate(zoneFQDN)
	return msg
}

// sendUpdate 发送DNS更新消息
func (c *DNSClient) sendUpdate(msg *dns.Msg) error {
	// 设置TSIG认证
	msg.SetTsig(c.tsigKeyName, dns.HmacSHA256, 300, time.Now().Unix())

	serverAddr := fmt.Sprintf("%s:%s", c.Server, c.Port)
	response, _, err := c.client.Exchange(msg, serverAddr)
	if err != nil {
		return fmt.Errorf("DNS exchange failed: %v", err)
	}

	if response.MsgHdr.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS update failed with RCODE: %s",
			dns.RcodeToString[response.MsgHdr.Rcode])
	}

	return nil
}

// AddRecord 添加解析记录
func (c *DNSClient) AddRecord(zone, record, rtype, value string, ttl uint32) error {
	msg := c.createUpdateMessage(zone)

	recordFQDN := dns.Fqdn(record)

	// 创建记录
	addRR := fmt.Sprintf("%s %d %s %s", recordFQDN, ttl, rtype, value)
	rrAdd, err := dns.NewRR(addRR)
	if err != nil {
		return fmt.Errorf("failed to create RR: %v", err)
	}

	// 添加记录
	msg.Insert([]dns.RR{rrAdd})

	return c.sendUpdate(msg)
}

// DeleteSpecificRecord 删除特定的记录（通过完整的记录信息）
func (c *DNSClient) DeleteSpecificRecord(zone, record, rtype, value string, ttl uint32) error {
	msg := c.createUpdateMessage(zone)
	recordFQDN := dns.Fqdn(record)

	// 创建完整记录用于删除
	deleteRR := fmt.Sprintf("%s %d %s %s", recordFQDN, ttl, rtype, value)
	rrDel, err := dns.NewRR(deleteRR)
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}

	// 删除特定记录
	msg.Remove([]dns.RR{rrDel})
	return c.sendUpdate(msg)
}

// DeleteRecordByType 删除指定名称和类型的所有记录
func (c *DNSClient) DeleteRecordByType(zone, record, rtype string, ttl uint32) error {
	msg := c.createUpdateMessage(zone)

	recordFQDN := dns.Fqdn(record)

	// 创建删除记录模板（只指定类型）
	deleteRR := fmt.Sprintf("%s %d %s", recordFQDN, ttl, rtype)
	rrDel, err := dns.NewRR(deleteRR)
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}

	// 删除指定类型的所有记录
	msg.RemoveRRset([]dns.RR{rrDel})

	return c.sendUpdate(msg)
}

// DeleteRecordByName 删除指定名称的所有记录
func (c *DNSClient) DeleteRecordByName(zone, record string) error {
	msg := c.createUpdateMessage(zone)

	recordFQDN := dns.Fqdn(record)

	// 创建删除记录模板（使用ANY类型）
	deleteRRStr := fmt.Sprintf("%s 0 ANY", recordFQDN)
	deleteRR, err := dns.NewRR(deleteRRStr)
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}

	// 删除记录
	msg.RemoveRRset([]dns.RR{deleteRR})

	return c.sendUpdate(msg)
}

// UpdateRecord 修改解析记录（先删除后添加）
func (c *DNSClient) UpdateRecord(zone, record, rtype, oldvalue, newvalue string, ttl uint32) error {
	msg := c.createUpdateMessage(zone)

	recordFQDN := dns.Fqdn(record)

	// 删除现有记录
	deleteRRStr := fmt.Sprintf("%s %d %s %s", recordFQDN, ttl, rtype, oldvalue)
	deleteRR, err := dns.NewRR(deleteRRStr)
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}
	msg.Remove([]dns.RR{deleteRR})

	// 添加新记录
	addRRStr := fmt.Sprintf("%s %d %s %s", recordFQDN, ttl, rtype, newvalue)
	addRR, err := dns.NewRR(addRRStr)
	if err != nil {
		return fmt.Errorf("failed to create add RR: %v", err)
	}
	msg.Insert([]dns.RR{addRR})

	return c.sendUpdate(msg)
}

// RemoveAllRRset 删除所有记录
func (c *DNSClient) RemoveAllRRset(zone string) error {
	// 先获取所有记录
	records, err := c.FetchRecords(zone)
	if err != nil {
		return fmt.Errorf("failed to fetch records: %v", err)
	}

	// 创建更新消息
	msg := c.createUpdateMessage(zone)

	// 逐个删除记录
	var removeRRs []dns.RR
	for _, record := range records {
		recordFQDN := dns.Fqdn(record.Name)
		rrStr := fmt.Sprintf("%s %d %s %s", recordFQDN, record.TTL, record.Type, record.Value)
		rr, err := dns.NewRR(rrStr)
		if err != nil {
			return fmt.Errorf("failed to create RR for %s: %v", record.Name, err)
		}
		removeRRs = append(removeRRs, rr)
	}

	msg.Remove(removeRRs)
	return c.sendUpdate(msg)
}

// SyncAddRecords 同步解析记录（替换所有记录）
func (c *DNSClient) SyncAddRecords(zone string, records []DNSRecord) error {
	msg := c.createUpdateMessage(zone)

	// // 删除区域中的所有记录
	// zoneFQDN := dns.Fqdn(zone)
	// deleteAllRRStr := fmt.Sprintf("%s 0 ANY", zoneFQDN)
	// deleteAllRR, err := dns.NewRR(deleteAllRRStr)
	// if err != nil {
	// 	return fmt.Errorf("failed to create delete all RR: %v", err)
	// }
	// msg.RemoveRRset([]dns.RR{deleteAllRR})

	// 添加所有新记录
	var addRRs []dns.RR
	for _, record := range records {
		recordFQDN := dns.Fqdn(record.Name)
		rrStr := fmt.Sprintf("%s %d %s %s", recordFQDN, record.TTL, record.Type, record.Value)
		rr, err := dns.NewRR(rrStr)
		if err != nil {
			return fmt.Errorf("failed to create RR for %s: %v", record.Name, err)
		}
		addRRs = append(addRRs, rr)
	}

	msg.Insert(addRRs)

	return c.sendUpdate(msg)
}

// FetchRecords 从服务器获取区域的所有记录（服务器端同步到本地）
func (c *DNSClient) FetchRecords(zone string) ([]DNSRecord, error) {
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
func (c *DNSClient) parseRR(rr dns.RR) *DNSRecord {
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
	case *dns.SRV:
		record.Type = "SRV"
		record.Value = fmt.Sprintf("%d %d %d %s", r.Priority, r.Weight, r.Port, strings.TrimSuffix(r.Target, "."))
	case *dns.CAA:
		record.Type = "CAA"
		record.Value = fmt.Sprintf("%d %s \"%s\"", r.Flag, r.Tag, r.Value)
	case *dns.DNSKEY:
		record.Type = "DNSKEY"
		record.Value = fmt.Sprintf("%d %d %d %s", r.Flags, r.Protocol, r.Algorithm, r.PublicKey)
	case *dns.DS:
		record.Type = "DS"
		record.Value = fmt.Sprintf("%d %d %d %s", r.KeyTag, r.Algorithm, r.DigestType, r.Digest)
	case *dns.NAPTR:
		record.Type = "NAPTR"
		record.Value = fmt.Sprintf("%d %d \"%s\" \"%s\" \"%s\" %s",
			r.Order, r.Preference, r.Flags, r.Service, r.Regexp, strings.TrimSuffix(r.Replacement, "."))
	case *dns.SSHFP:
		record.Type = "SSHFP"
		record.Value = fmt.Sprintf("%d %d %s", r.Algorithm, r.Type, r.FingerPrint)
	case *dns.TLSA:
		record.Type = "TLSA"
		record.Value = fmt.Sprintf("%d %d %d %s", r.Usage, r.Selector, r.MatchingType, r.Certificate)
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

// SyncZoneToLocal 从服务器同步到本地（服务器端同步到本地）
func (c *DNSClient) SyncZoneToLocal(zone string) ([]DNSRecord, error) {
	fmt.Printf("正在从服务器 %s 同步区域 %s 的记录...\n", c.Server, zone)

	records, err := c.FetchRecords(zone)
	if err != nil {
		return nil, fmt.Errorf("从服务器获取记录失败: %v", err)
	}

	fmt.Printf("成功同步 %d 条记录\n", len(records))
	return records, nil
}

// SyncZoneToServer 将本地数据同步到服务器（同步本地数据到服务器）
func (c *DNSClient) SyncZoneToServer(zone string, records []DNSRecord) error {
	fmt.Printf("正在同步 %d 条记录到服务器 %s 的区域 %s ...\n", len(records), c.Server, zone)

	err := c.SyncAddRecords(zone, records)
	if err != nil {
		return fmt.Errorf("同步记录到服务器失败: %v", err)
	}

	fmt.Println("记录同步成功")
	return nil
}

// CompareRecords 比较两组记录的差异
func (c *DNSClient) CompareRecords(local, remote []DNSRecord) *SyncDiff {
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

// BidirectionalSync 双向同步（先从服务器同步到本地，再同步本地到服务器）
func (c *DNSClient) BidirectionalSync(zone string, localRecords []DNSRecord) ([]DNSRecord, error) {
	fmt.Printf("开始双向同步区域 %s ...\n", zone)

	// 1. 从服务器获取当前记录
	remoteRecords, err := c.FetchRecords(zone)
	if err != nil {
		return nil, fmt.Errorf("获取服务器记录失败: %v", err)
	}

	// 2. 比较差异
	diff := c.CompareRecords(localRecords, remoteRecords)
	diff.PrintDiff()
	fmt.Println(localRecords)
	// 3. 同步本地记录到服务器
	err = c.SyncAddRecords(zone, localRecords)
	if err != nil {
		return nil, fmt.Errorf("同步到服务器失败: %v", err)
	}

	fmt.Println("双向同步完成")
	return remoteRecords, nil
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
