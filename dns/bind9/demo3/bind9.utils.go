package demo3

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSClient DNS客户端
type DNSClient struct {
	Server      string      // DNS服务器地址
	KeyName     string      // TSIG密钥名称
	Secret      string      // TSIG密钥值
	Port        string      // DNS服务器端口
	client      *dns.Client // DNS客户端实例
	tsigKeyName string      // 格式化后的TSIG密钥名称
}

// DNSRecord DNS记录结构
type DNSRecord struct {
	DomainName string // 域名名称
	RR         string // 主机记录
	Type       string // 解析记录类型
	Value      string // 记录值
	TTL        uint32 // 解析生效时间
}

// SyncDiff 同步差异结构
type SyncDiff struct {
	Remote []DNSRecord // 仅存在于远程记录
	Local  []DNSRecord // 仅存在于本地记录
}

/*
	NewDNSClient 创建一个新的DNS客户端实例

参数:

	server: DNS服务器地址
	keyName: TSIG密钥名称
	secret: TSIG密钥密文
	port: DNS服务器端口，如果为空则默认使用53端口

返回值:

	指向DNSClient实例的指针
*/
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

/*
dnsStringToType 将DNS记录类型字符串转换为对应的DNS类型码
参数:

	rtype: DNS记录类型字符串，如"A", "AAAA", "CNAME"等

返回值:

	uint16: 对应的DNS类型码，如果未找到匹配的类型则返回dns.TypeNone
*/
// dns.StringToType[rtype]
// func dnsStringToType(rtype string) uint16 {
// 	switch rtype {
// 	case "A":
// 		// A - IPv4地址记录
// 		return dns.TypeA
// 	case "AAAA":
// 		// AAAA - IPv6地址记录
// 		return dns.TypeAAAA
// 	case "CNAME":
// 		// CNAME - 规范名称记录，用于域名别名
// 		return dns.TypeCNAME
// 	case "MX":
// 		// MX - 邮件交换记录，指定邮件服务器
// 		return dns.TypeMX
// 	case "NS":
// 		// NS - 名称服务器记录，指定域名的DNS服务器
// 		return dns.TypeNS
// 	case "PTR":
// 		// PTR - 指针记录，用于反向DNS查找
// 		return dns.TypePTR
// 	case "TXT":
// 		// TXT - 文本记录
// 		return dns.TypeTXT
// 	case "SRV":
// 		// SRV - 服务定位记录
// 		return dns.TypeSRV
// 	case "CAA":
// 		// CAA - 证书颁发机构授权记录
// 		return dns.TypeCAA
// 	case "DNSKEY":
// 		// DNSKEY - DNS密钥记录
// 		return dns.TypeDNSKEY
// 	case "DS":
// 		// DS - 委派签名记录
// 		return dns.TypeDS
// 	case "NAPTR":
// 		// NAPTR - 命名权威指针记录
// 		return dns.TypeNAPTR
// 	case "SSHFP":
// 		// SSHFP - SSH公钥指纹记录
// 		return dns.TypeSSHFP
// 	case "TLSA":
// 		// TLSA - TLSA证书关联记录
// 		return dns.TypeTLSA
// 	default:
// 		return dns.TypeNone
// 	}
// }

/*
	createUpdateMessage 创建一个用于DNS更新的空消息对象

该函数会将给定的zone名称转换为FQDN格式，并初始化一个DNS消息对象用于UPDATE操作

参数:
  - zone: 需要更新的DNS区域名称

返回值:
  - *dns.Msg: 初始化完成的DNS消息对象，已设置为UPDATE模式
*/
func (c *DNSClient) createUpdateMessage(zone string) *dns.Msg {
	zoneFQDN := dns.Fqdn(zone) // 将区域名称转换为完全限定域名(FQDN)格式
	msg := new(dns.Msg)        // 创建新的DNS消息对象
	msg.SetUpdate(zoneFQDN)    // 设置消息为UPDATE操作模式，指定操作区域
	return msg                 // 返回初始化完成的消息对象
}

/*
	sendUpdate 向DNS服务器发送更新消息

参数:

	msg - 要发送的DNS消息

返回值:

	error - 发送过程中出现的错误，nil表示成功
*/
func (c *DNSClient) sendUpdate(msg *dns.Msg) error {
	// 设置TSIG认证
	msg.SetTsig(c.tsigKeyName, dns.HmacSHA256, 300, time.Now().Unix())

	// 构造服务器地址并发送DNS查询
	serverAddr := fmt.Sprintf("%s:%s", c.Server, c.Port)
	response, _, err := c.client.Exchange(msg, serverAddr)
	if err != nil {
		return fmt.Errorf("DNS exchange failed: %v", err)
	}

	// 检查DNS响应状态码
	if response.MsgHdr.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS update failed with RCODE: %s",
			dns.RcodeToString[response.MsgHdr.Rcode])
	}

	return nil
}

/*
	parseRR 将给定的 DNS 资源记录（RR）解析为统一的 DNSRecord 结构。

该函数会根据资源记录的类型提取相应的字段，并进行格式化处理，以保证数据的一致性。

参数:
  - rr: dns.RR 类型，表示一个 DNS 资源记录。

返回值:
  - *DNSRecord: 解析后的 DNS 记录结构体指针；如果无法解析，则返回 nil。
*/
func (c *DNSClient) parseRR(rr dns.RR) *DNSRecord {
	header := rr.Header()

	// 移除末尾的点以保持一致性
	name := strings.TrimSuffix(header.Name, ".")

	record := &DNSRecord{
		RR:   name,
		Type: dns.TypeToString[header.Rrtype],
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

/*
	CompareRecords 比较本地和远程DNS记录的差异

该函数通过比对本地和远程两组DNS记录，找出需要同步的记录

参数:
  - local: 本地DNS记录列表
  - remote: 远程DNS记录列表

返回值:
  - *SyncDiff: 包含差异记录的结构体指针
  - Local: 仅存在于本地的记录列表（需要添加到远程）
  - Remote: 仅存在于远程的记录列表（需要添加到本地）
*/
func (c *DNSClient) CompareRecords(local, remote []DNSRecord) *SyncDiff {
	// 初始化差异结构体
	diff := &SyncDiff{
		Local:  []DNSRecord{},
		Remote: []DNSRecord{},
	}

	// 将记录转换为map便于比较
	localMap := make(map[string]DNSRecord)
	remoteMap := make(map[string]DNSRecord)

	// 构建本地记录映射
	for _, record := range local {
		// 使用记录的名称、类型和值作为唯一键
		key := fmt.Sprintf("%s|%s|%s", record.RR, record.Type, record.Value)
		localMap[key] = record
	}

	// 构建远程记录映射
	for _, record := range remote {
		// 使用记录的名称、类型和值作为唯一键
		key := fmt.Sprintf("%s|%s|%s", record.RR, record.Type, record.Value)
		remoteMap[key] = record
	}

	// 找出需要添加的记录（在本地但不在远程）
	for key, record := range localMap {
		if _, exists := remoteMap[key]; !exists {
			diff.Local = append(diff.Local, record)
		}
	}

	// 找出需要删除的记录（在远程但不在本地）
	for key, record := range remoteMap {
		if _, exists := localMap[key]; !exists {
			diff.Remote = append(diff.Remote, record)
		}
	}

	return diff
}

/*
	CompareDiff 比较两个DNS记录列表的差异，返回在first中存在但不在second中的记录

参数:

	first - 用于比较的基准
	second - 需要检查缺失元素

返回值:

	*[]DNSRecord - 指向差异记录列表的指针，包含在first中存在但不在second中的记录
*/
func (c *DNSClient) CompareDiff(first, second []DNSRecord) []DNSRecord {
	diff := []DNSRecord{}
	// 将记录转换为map便于比较
	firstMap := make(map[string]DNSRecord)
	secondMap := make(map[string]DNSRecord)

	// 遍历first切片中的每条记录，以记录的名称、类型和值构建唯一键，并存储到firstMap中
	// key的格式为"名称|类型|值"，确保相同内容的记录会被去重
	for _, record := range first {
		// 使用记录的名称、类型和值作为唯一键
		key := fmt.Sprintf("%s|%s|%s", record.RR, record.Type, record.Value)
		firstMap[key] = record
	}
	// 遍历second切片，将每条记录转换为唯一键并存储到secondMap中
	// 以记录的名称、类型和值组合作为唯一键，避免重复记录
	for _, record := range second {
		// 使用记录的名称、类型和值作为唯一键
		key := fmt.Sprintf("%s|%s|%s", record.RR, record.Type, record.Value)
		secondMap[key] = record
	}
	// 遍历firstMap，找出在secondMap中不存在的记录
	for key, record := range firstMap {
		if _, exists := secondMap[key]; !exists {
			diff = append(diff, record)
		}
	}
	return diff
}
