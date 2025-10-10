# bind9

## nsupdate 配置
https://www.cnblogs.com/chengd/articles/14963484.html
https://www.jianshu.com/p/0a47b6c3d5e0
https://github.com/SpComb/go-nsupdate

## 安装bind9
```bash
yum -y install bind bind-utils
rpm -aq bind
```
```bash
# 配置
# rpm 包
sed -i \
-e "s#listen-on port 53.*#listen-on port 53 { any; };#" \
-e "s#allow-query.*#allow-query     { any; };#" \
-e "/options/a\        forwarders { 101.226.4.6; };" \
-e "/options/a\        allow-update { key "dnsadmin"; };" \
-e "/options/a\        allow-transfer { key "dnsadmin"; };" \
/etc/named.conf

cat >> /etc/named.conf << "EOF"
include "/etc/named/named.conf.dnsadmin";
EOF
```
## 配置tsig
> ==需确认配置文件中以下配置是开启状态==
```bash
options {
    dnssec-enable yes;
    dnssec-validation yes;
}
```

```bash
cat > /etc/named/named.conf.dnsadmin << "EOF"
include "/etc/named/named.conf.dnsadmin.tsig-keygen";
include "/etc/named/named.conf.dnsadmin.acl";
include "/etc/named/named.conf.dnsadmin.zones";
EOF

tsig-keygen -a hmac-sha256 dnsadmin >> /etc/named/named.conf.dnsadmin.tsig-keygen

cat >> /etc/named/named.conf.dnsadmin.acl << "EOF"
acl "dns-updaters" {
    localhost;
    10.10.14.0/24;
};
EOF
cat >> /etc/named/named.conf.dnsadmin.zones << "EOF"
zone "demo.com" {
    type master;
    file "/var/named/data/demo.com.zone";
};
EOF

cat > /var/named/data/demo.com.zone << EOF
\$ORIGIN .
\$TTL 600        ; 10 minutes
demo.com             IN SOA  dns. dnsadmin. (
                                $(date +"%Y%m%d%H") ; serial
                                10800      ; refresh (3 hours)
                                900        ; retry (15 minutes)
                                604800     ; expire (1 week)
                                86400      ; minimum (1 day)
                                )
                        NS      dns.
EOF
chown -R root.named /etc/named
chown -R named.named /var/named/data/
chmod 644 /etc/named/*

named-checkconf /etc/named.conf
named-checkconf /etc/named/named.conf.dnsadmin
named-checkconf /etc/named/named.conf.dnsadmin.acl
named-checkconf /etc/named/named.conf.dnsadmin.tsig-keygen
named-checkconf /etc/named/named.conf.dnsadmin.zones
named-checkzone demo.com /var/named/data/demo.com.zone

systemctl enable --now named.service
```

## 获取密钥
```bash
KEY_FILE="/etc/named/named.conf.dnsadmin.tsig-keygen"
KEY_NAME=$(sed -n 's/^key "\([^"]*\)".*/\1/p' "$KEY_FILE")
KEY_SECRET=$(sed -n 's/.*secret "\([^"]*\)".*/\1/p' "$KEY_FILE")
KEY_ALGORITHM=$(sed -n 's/.*algorithm \([^;]*\);.*/\1/p' "$KEY_FILE")
echo "KEY_NAME: $KEY_NAME"
echo "KEY_SECRET: $KEY_SECRET"
echo "KEY_ALGORITHM: $KEY_ALGORITHM"
```

## 测试
```bash
rndc sync -clean

cat > /tmp/test1 << EOF
key ${KEY_NAME} ${KEY_SECRET}
server 127.0.0.1
zone demo.com
update add test.demo.com. 300 A 192.168.1.200
update add test.demo.com. 300 TXT "Test record"
send
EOF
nsupdate /tmp/test1

cat > /tmp/test2 << EOF
server 127.0.0.1
zone demo.com
update add test.demo.com. 300 A 192.168.1.200
update add test.demo.com. 300 TXT "Test record"
send
EOF
nsupdate -y ${KEY_ALGORITHM}:${KEY_NAME}:${KEY_SECRET} /tmp/test2

cat > /tmp/test3 << EOF
server 10.10.14.20
zone demo.com
update add mail.demo.com. 300 A 192.168.1.200
update add demo.com. 300 MX 10 mail.demo.com.
send
EOF
nsupdate -y ${KEY_ALGORITHM}:${KEY_NAME}:${KEY_SECRET} /tmp/test3


```
## 配置 TSIG (旧)
https://www.cnblogs.com/liujiaxin2018/p/14125665.html
```bash
## -a 指定加密算法，-b 指定秘钥长度 ，-n 指定主机类型， 执行命令后会在当前目录生成公钥和私钥文件
# dnssec-keygen -a HMAC-MD5 -b 128 -n  HOST master-slave


tsig-keygen -a hmac-sha256 mykeyname >> /etc/bind/named.conf.local


cat >> /etc/bind/named.conf.local << "EOF"
acl "dns-updaters" {
    localhost;
    192.168.1.0/24;
    10.0.0.0/8;
};
EOF
cat >> /etc/bind/named.conf.default-zones << "EOF"
zone "demo.com" {
    type master;
    file "/etc/bind/db.demo.com";
    allow-update { key "mykeyname"; };
};
EOF


named-checkconf /etc/bind/named.conf
named-checkconf /etc/bind/named.conf.local

named-checkzone demo.com /etc/bind/db.demo.com

systemctl restart bind9
```


allow-transfer { key mykeyname; };

allow-update { key mykeyname;};

key "mykeyname" {
        algorithm hmac-sha256;
        secret "0cs/pBJacfUBrbHRnu/RbCtHIZiayBRxpz/NtLjmeiE=";
};