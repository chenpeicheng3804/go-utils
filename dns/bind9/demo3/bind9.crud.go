package demo3

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

/*
	DescribeDomainRecords 通过AXFR区域传输获取指定域名的所有DNS记录

参数:

	zone: 要查询的域名区域

返回值:

	[]DNSRecord: DNS记录列表
	error: 操作成功返回nil，失败返回错误信息
*/
func (c *DNSClient) DescribeDomainRecords(zone string) ([]DNSRecord, error) {
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
			// // 跳过SOA记录（通常由服务器管理）
			// if _, ok := rr.(*dns.SOA); ok {
			// 	continue
			// }
			// 跳过系统管理的记录类型
			rrType := rr.Header().Rrtype
			if rrType == dns.TypeSOA || rrType == dns.TypeNS {
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

/*
	QueryRecord 查询DNS记录

参数:

	name: 要查询的域名
	rtype: 记录类型，如"A", "AAAA", "MX"等

返回值:

	[]dns.RRDNS: 资源记录切片
	error: 操作成功返回nil，失败返回错误信息
*/
func (d *DNSClient) QueryRecord(name string, rtype string) ([]dns.RR, error) {
	// 创建DNS客户端和消息对象
	c := new(dns.Client)
	m := new(dns.Msg)

	// 设置查询问题部分，将域名转换为FQDN格式并设置记录类型
	m.SetQuestion(dns.Fqdn(name), dns.StringToType[rtype])
	m.RecursionDesired = true

	// 向DNS服务器发送查询请求
	r, _, err := c.Exchange(m, d.Server+":53")
	if err != nil {
		return nil, err
	}

	// 返回查询结果中的答案部分
	return r.Answer, nil
}

/*
	AddDomainRecord 向DNS区域添加域名记录

参数:

	zone: DNS区域名称
	record: 要添加的记录名称
	rtype: 记录类型(如A、CNAME等)
	value: 记录值
	ttl: 记录的生存时间

返回值:

	error: 操作成功返回nil，失败返回错误信息
*/
func (c *DNSClient) AddDomainRecord(zone, record, rtype, value string, ttl uint32) error {
	if ttl == 0 {
		ttl = 600
	}
	// 创建DNS更新消息
	msg := c.createUpdateMessage(zone)

	// 构造完整的记录域名
	recordFQDN := dns.Fqdn(record)

	// 兼容TXT
	var addRR string
	txtValue := value
	if rtype == "TXT" {
		if !strings.HasPrefix(txtValue, "\"") {
			txtValue = "\"" + txtValue + "\""
		}
	}

	// 创建记录
	addRR = fmt.Sprintf("%s %d %s %s", recordFQDN, ttl, rtype, txtValue)
	//fmt.Println(addRR)
	rrAdd, err := dns.NewRR(addRR)
	//fmt.Println(rrAdd)
	if err != nil {
		return fmt.Errorf("failed to create RR: %v", err)
	}

	// 添加记录到更新消息中
	msg.Insert([]dns.RR{rrAdd})

	// 发送更新请求
	return c.sendUpdate(msg)
}

/*
	UpdateDomainRecord 更新DNS域名记录

该函数通过DNS UPDATE协议更新指定域名的记录，先删除旧记录再添加新记录
参数:

	zone: DNS区域名称
	record: 要更新的记录名称
	rtype: 记录类型（如A、CNAME等）
	oldvalue: 旧记录值
	newvalue: 新记录值
	ttl: 记录的生存时间

返回值:

	error: 操作过程中发生的错误，如果操作成功则返回nil
*/
func (c *DNSClient) UpdateDomainRecord(zone, record, rtype, oldvalue, newvalue string, ttl uint32) error {
	if ttl == 0 {
		ttl = 600
	}
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

/*
	DeleteDomainRecord 从指定的DNS区域中删除域名记录

参数:

	zone: DNS区域名称
	record: 要删除的记录名称
	rtype: 记录类型(如A、CNAME等)
	value: 记录值
	ttl: 记录的生存时间

返回值:

	error: 删除操作的错误信息，成功时返回nil
*/
func (c *DNSClient) DeleteDomainRecord(zone, record, rtype, value string, ttl uint32) error {
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

/*
DeleteRecordByType 根据记录类型删除指定记录的所有值

该函数通过DNS UPDATE协议删除指定域名和记录类型的所有记录，
与删除特定记录不同，此函数会删除该类型下的所有记录。

参数:

	zone: DNS区域名称（例如"example.com"）
	record: 要删除记录的域名（例如"www.example.com"）
	rtype: 要删除的记录类型（例如"A"、"AAAA"、"CNAME"等）

返回值:

	error: 删除操作过程中发生的错误，如果删除成功则返回nil

示例:

	// 删除example.com区域内所有www.example.com的A记录
	err := client.DeleteRecordByType("example.com", "www.example.com", "A")
*/
func (c *DNSClient) DeleteRecordByType(zone, record, rtype string) error {
	// 创建DNS更新消息
	msg := c.createUpdateMessage(zone)

	// 将记录名称转换为完全限定域名(FQDN)
	recordFQDN := dns.Fqdn(record)

	// 直接使用 dns.NewRR 构造记录，无需指定 TTL
	rrDel, err := dns.NewRR(fmt.Sprintf("%s %s", recordFQDN, rtype))
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}

	// 使用RemoveRRset删除指定记录名和类型的所有记录
	msg.RemoveRRset([]dns.RR{rrDel})

	// 发送DNS更新消息
	return c.sendUpdate(msg)
}

/*
DeleteRecordByName 根据记录名称删除指定DNS记录

该函数通过DNS UPDATE协议删除指定域名的所有记录（不区分记录类型），
使用ANY类型匹配指定记录名下的所有记录类型。

参数:
  - zone: DNS区域名称（例如"example.com"）
  - record: 要删除记录的域名（例如"www.example.com"，不需要FQDN格式）

返回值:
  - error: 删除操作过程中发生的错误，如果删除成功则返回nil

示例:

	// 删除example.com区域内所有www.example.com的记录（包括A、AAAA、CNAME等所有类型）
	err := client.DeleteRecordByName("example.com", "www.example.com")
*/
func (c *DNSClient) DeleteRecordByName(zone, record string) error {
	// 创建DNS更新消息
	msg := c.createUpdateMessage(zone)

	// 将记录名称转换为完全限定域名(FQDN)
	recordFQDN := dns.Fqdn(record)

	// 创建删除记录模板（使用ANY类型匹配所有记录类型）
	deleteRRStr := fmt.Sprintf("%s 0 ANY", recordFQDN)
	deleteRR, err := dns.NewRR(deleteRRStr)
	if err != nil {
		return fmt.Errorf("failed to create delete RR: %v", err)
	}

	// 使用RemoveRRset删除指定记录名的所有记录
	msg.RemoveRRset([]dns.RR{deleteRR})

	// 发送DNS更新消息
	return c.sendUpdate(msg)
}

/*
	DeleteDomainAll 删除指定区域的所有DNS记录

参数:

	zone: 要删除记录的DNS区域名称

返回值:

	error: 删除操作过程中发生的错误，如果删除成功则返回nil
*/
func (c *DNSClient) DeleteDomainAll(zone string) error {
	// 先获取所有记录
	records, err := c.DescribeDomainRecords(zone)
	if err != nil {
		return fmt.Errorf("failed to fetch records: %v", err)
	}

	// 创建更新消息
	msg := c.createUpdateMessage(zone)

	// 逐个删除记录
	var removeRRs []dns.RR
	for _, record := range records {
		recordFQDN := dns.Fqdn(record.RR)
		var rr dns.RR
		var err error

		// 对于 TXT 记录进行特殊处理
		if record.Type == "TXT" {
			// TXT 记录值需要正确格式化
			rrStr := fmt.Sprintf("%s %d %s \"%s\"", recordFQDN, record.TTL, record.Type, record.Value)
			rr, err = dns.NewRR(rrStr)
		} else {
			rrStr := fmt.Sprintf("%s %d %s %s", recordFQDN, record.TTL, record.Type, record.Value)
			rr, err = dns.NewRR(rrStr)
		}

		if err != nil {
			// 如果解析失败，尝试使用 RemoveRRset 方法删除该记录类型的所有记录
			fallbackRRStr := fmt.Sprintf("%s %s", recordFQDN, record.Type)
			fallbackRR, fallbackErr := dns.NewRR(fallbackRRStr)
			if fallbackErr != nil {
				return fmt.Errorf("failed to create RR for %s: %v", record.RR, err)
			}
			msg.RemoveRRset([]dns.RR{fallbackRR})
			continue
		}
		removeRRs = append(removeRRs, rr)
	}

	if len(removeRRs) > 0 {
		msg.Remove(removeRRs)
	}
	return c.sendUpdate(msg)
}
