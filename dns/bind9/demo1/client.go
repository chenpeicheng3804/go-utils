// client.go
package demo1

import (
	"fmt"
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
	rrStr := fmt.Sprintf("%s %d %s %s", recordFQDN, ttl, rtype, value)
	rr, err := dns.NewRR(rrStr)
	if err != nil {
		return fmt.Errorf("failed to create RR: %v", err)
	}

	// 添加记录
	msg.Insert([]dns.RR{rr})

	return c.sendUpdate(msg)
}

// DeleteRecord 删除解析记录
func (c *DNSClient) DeleteRecord(zone, record, rtype string, ttl uint32) error {
	msg := c.createUpdateMessage(zone)

	recordFQDN := dns.Fqdn(record)

	// 创建删除记录模板
	rrStr := fmt.Sprintf("%s %d %s", recordFQDN, ttl, rtype)
	rr, err := dns.NewRR(rrStr)
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}

	// 删除记录
	msg.RemoveRRset([]dns.RR{rr})

	return c.sendUpdate(msg)
}

// DeleteRecordByName 删除指定名称的所有记录
func (c *DNSClient) DeleteRecordByName(zone, record string) error {
	msg := c.createUpdateMessage(zone)

	recordFQDN := dns.Fqdn(record)

	// 创建删除记录模板（不指定类型）
	rrStr := fmt.Sprintf("%s 0 ANY", recordFQDN)
	rr, err := dns.NewRR(rrStr)
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}

	// 删除记录
	msg.RemoveRRset([]dns.RR{rr})

	return c.sendUpdate(msg)
}

// UpdateRecord 修改解析记录（先删除后添加）
func (c *DNSClient) UpdateRecord(zone, record, rtype, value string, ttl uint32) error {
	msg := c.createUpdateMessage(zone)

	recordFQDN := dns.Fqdn(record)

	// 删除现有记录
	deleteRRStr := fmt.Sprintf("%s %d %s", recordFQDN, ttl, rtype)
	deleteRR, err := dns.NewRR(deleteRRStr)
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}
	msg.RemoveRRset([]dns.RR{deleteRR})

	// 添加新记录
	addRRStr := fmt.Sprintf("%s %d %s %s", recordFQDN, ttl, rtype, value)
	addRR, err := dns.NewRR(addRRStr)
	if err != nil {
		return fmt.Errorf("failed to create add RR: %v", err)
	}
	msg.Insert([]dns.RR{addRR})

	return c.sendUpdate(msg)
}

// SyncRecords 同步解析记录（替换所有记录）
func (c *DNSClient) SyncRecords(zone string, records []DNSRecord) error {
	msg := c.createUpdateMessage(zone)

	// 删除区域中的所有记录
	zoneFQDN := dns.Fqdn(zone)
	deleteAllRRStr := fmt.Sprintf("%s 0 ANY", zoneFQDN)
	deleteAllRR, err := dns.NewRR(deleteAllRRStr)
	if err != nil {
		return fmt.Errorf("failed to create delete all RR: %v", err)
	}
	msg.RemoveRRset([]dns.RR{deleteAllRR})

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

// DNSRecord DNS记录结构
type DNSRecord struct {
	Name  string
	Type  string
	Value string
	TTL   uint32
}
