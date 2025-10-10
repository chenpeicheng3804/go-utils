package demo3

import (
	"fmt"
	"github.com/miekg/dns"
)

// 单向差异Zone同步
// 将本地Zone数据同步到服务器
func (c *DNSClient) SyncZoneToServer(zone string, local []DNSRecord) ([]DNSRecord, error) {
	msg := c.createUpdateMessage(zone)
	// 获取服务器Zone数据
	remote, err := c.DescribeDomainRecords(zone)
	if err != nil {
		return nil, err
	}

	// 获取服务器中不存在的数据
	diff := c.CompareDiff(remote, local)
	var addRRs []dns.RR
	for _, record := range diff {
		recordFQDN := dns.Fqdn(record.RR)
		rrStr := fmt.Sprintf("%s %d %s %s", recordFQDN, record.TTL, record.Type, record.Value)
		rr, err := dns.NewRR(rrStr)
		if err != nil {
			return nil, fmt.Errorf("failed to create RR for %s: %v", record.RR, err)
		}
		addRRs = append(addRRs, rr)
	}
	msg.Insert(addRRs)
	return diff, c.sendUpdate(msg)
}

// 将服务器Zone数据同步到本地
func (c *DNSClient) SyncZoneToLocal(zone string, local []DNSRecord) ([]DNSRecord, error) {
	// 获取服务器Zone数据
	remote, err := c.DescribeDomainRecords(zone)
	if err != nil {
		return nil, err
	}
	// 返回本地中不存在的数据
	return c.CompareDiff(local, remote), nil
}

// 双向差异Zone同步
