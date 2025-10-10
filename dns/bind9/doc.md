

### 创建新的DNS客户端

```go
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
```

###









