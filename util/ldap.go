package util

import (
	"errors"
	"fmt"

	"github.com/chenpeicheng3804/go-utils/util/log"
	"github.com/go-ldap/ldap/v3"
)

// LdapConfig ldap配置
type LdapConfig struct {
	Uri        string
	MainDomain string
	Dn         string
	Auth       Auth
}

// Auth认证信息
type Auth struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// ldapConn ldap连接
type ldapConn struct {
	dn         string
	mainDomain string
	conn       *ldap.Conn
}

// NewLdapConn
// 创建ldap连接
func NewLdapConn(config LdapConfig) (l *ldapConn, err error) {
	conn, err := ldap.DialURL(config.Uri)
	if err != nil {
		// log.Panicln(err)
		log.Debug().Err(err).Msg("")
		return nil, err
	}
	_, err = conn.SimpleBind(&ldap.SimpleBindRequest{
		Username: fmt.Sprintf("cn=%s,%s", config.Auth.UserName, config.Dn),
		Password: config.Auth.Password,
	})

	if err != nil {
		// log.Panicln("ldap password is error: ", ldap.LDAPResultInvalidCredentials)
		log.Debug().Err(err).Msg("ldap password is error")
		return nil, err
	}
	return &ldapConn{conn: conn, dn: config.Dn, mainDomain: config.MainDomain}, nil
}

// Close 关闭ldap连接
func (l *ldapConn) Close() {
	log.Debug().Msg("关闭ldap连接")
	l.conn.Close()
}

// search
// 根据条件进行查询
/*
	condition: 查询条件
	查询组
	l.search("(&(objectClass=organizationalUnit)(ou=%s))", group, "cn",group)
	查询用户
	l.search("(&(objectClass=person)(cn=%s))", username, "cn",username)
	查询用户组成员
	l.search("(&(objectClass=groupOfUniqueNames)(cn=%s))", "test", "uniqueMember","cn=dddd,ou=People,dc=jcszfw,dc=com")
*/
func (l *ldapConn) search(objectClass, condition, attributes, result string) bool {
	//fmt.Println(fmt.Sprintf(objectClass, condition))
	searchRequest := ldap.NewSearchRequest(
		l.dn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf(objectClass, condition),
		[]string{attributes},
		nil,
	)
	sr, err := l.conn.Search(searchRequest)
	if err != nil {
		log.Debug().Str("服务器返回错误", err.Error()).Msg("")
		return false
	}

	if len(sr.Entries) == 0 {
		log.Debug().Str("查询结果不存在", condition).Msg("")
		return false
	}

	for _, entry := range sr.Entries {
		//entry.Print()
		//fmt.Printf("%s: %v\n", entry.DN, entry.GetAttributeValues(attributes))
		log.Debug().Str("查询内容", entry.DN).Strs(attributes, entry.GetAttributeValues(attributes)).Msg("")
		for _, v := range entry.GetAttributeValues(attributes) {
			//fmt.Println("v", v)
			//fmt.Println("result", result)
			if v == result {
				return true
			}
		}
	}
	return false
}

// SearchUser 查询用户
func (l *ldapConn) SearchUser(username string) bool {
	return l.search("(&(objectClass=person)(cn=%s))", username, "cn", username)
}

// SearchGroup 查询组
func (l *ldapConn) SearchGroup(group string) bool {
	return l.search("(&(objectClass=organizationalUnit)(ou=%s))", group, "ou", group)
}

// Add 添加资源
func (l *ldapConn) add(cn string, attributes []ldap.EntryAttribute) error {
	addRequest := ldap.NewAddRequest(fmt.Sprintf("%s,%s", cn, l.dn), nil)
	for _, attr := range attributes {
		addRequest.Attribute(attr.Name, attr.Values)
	}
	err := l.conn.Add(addRequest)
	if err != nil {
		log.Debug().
			Str("添加资源", cn+","+l.dn).
			Str("添加失败", err.Error()).Msg("")
		return err
	}
	log.Debug().Str("添加成功", cn+","+l.dn).Msg("")
	return nil
}

type LdapUser struct {
	Username   string
	Ou         string
	Zhusername string
	Password   string
	Group      LdapGroup
}
type LdapGroup struct {
	Cn      string
	Ou      string
	Members []string
}

// AddUser 添加用户
func (l *ldapConn) AddUser(user *LdapUser) error {

	// 判断用户组是否存在
	if !l.SearchGroup(user.Ou) {
		log.Debug().Str("分组不存在", user.Ou).Msg("进行创建")
		err := l.add(fmt.Sprintf("ou=%s", user.Ou), []ldap.EntryAttribute{
			{
				Name:   "objectClass",
				Values: []string{"organizationalUnit"},
			},
			{
				Name:   "ou",
				Values: []string{user.Ou},
			},
		})
		if err != nil {
			log.Debug().Str("创建用户组失败", err.Error()).Msg("")
			return err
		}
	}

	if user.Username == "" {
		// 中文转拼音
		user.Username = Pinyin(user.Zhusername)
	}

	// 判断用户是否存在
	if l.SearchUser(user.Username) {
		log.Debug().Str("用户已存在", user.Username).Msg("")
		l.addMemberUser(user)
		return errors.New("用户已存在")
	}

	if user.Password == "" {
		user.Password = RandomString(12)
	}
	attributes := []ldap.EntryAttribute{
		{
			Name:   "objectClass",
			Values: []string{"top", "inetOrgPerson"},
		},
		{
			Name:   "mail",
			Values: []string{fmt.Sprintf("%s@%s", user.Username, l.mainDomain)},
		},
		{
			Name:   "displayName",
			Values: []string{user.Zhusername},
		},
		{
			Name:   "sn",
			Values: []string{user.Username},
		},
		{
			Name:   "userPassword",
			Values: []string{user.Password},
		},
	}
	if err := l.add(fmt.Sprintf("cn=%s,ou=%s", user.Username, user.Ou), attributes); err != nil {
		log.Debug().Err(err).Msg("add user error")
		return errors.New("add user error")
	}
	log.Info().
		Str("用户", user.Username).
		Str("密码", user.Password).
		Msg("添加用户成功")
	return l.addMemberUser(user)
}

// addMember 添加组
func (l *ldapConn) addMember(cn, ou string) error {
	// 判断ou是否存在
	if !l.SearchGroup(ou) {
		log.Debug().Str("Ou不存在", ou).Msg("进行创建")
		err := l.add(fmt.Sprintf("ou=%s", ou), []ldap.EntryAttribute{
			{
				Name:   "objectClass",
				Values: []string{"organizationalUnit"},
			},
			{
				Name:   "ou",
				Values: []string{ou},
			},
		})
		if err != nil {
			log.Debug().Str("创建Ou失败", err.Error()).Msg("")
			return err
		}
	}
	// 判断用户组是否存在
	if l.search("(&(objectClass=groupOfUniqueNames)(cn=%s))", cn, "cn", cn) {
		log.Debug().Str("member已存在", cn).Msg("")
		return errors.New("member已存在")
	}
	attributes := []ldap.EntryAttribute{
		{
			Name:   "objectClass",
			Values: []string{"groupOfUniqueNames"},
		},
		{
			Name:   "uniquemember",
			Values: []string{""},
		},
	}
	return l.add(fmt.Sprintf("cn=%s,ou=%s", cn, ou), attributes)
}

// addMemberUser 用户添加用户组
func (l *ldapConn) addMemberUser(user *LdapUser) error {
	for _, m := range user.Group.Members {
		l.addMember(m, user.Group.Ou)
		// result := fmt.Sprintf("cn=%s,ou=%s,%s", user.Username, user.Ou, l.dn)
		result := fmt.Sprintf("cn=%s,ou=%s,%s", m, user.Group.Ou, l.dn)
		filter := fmt.Sprintf("(objectClass=person)(memberof=cn=%s,ou=%s,%s))", m, user.Group.Ou, l.dn)
		if l.search("(&(cn=%s)"+filter, user.Username, "memberOf", result) {
			log.Debug().Str("memberof", user.Username).Str("已存在", result).Msg("跳过")
			continue
		}
		// if l.search("(&(objectClass=groupOfUniqueNames)(cn=%s))", m, "uniqueMember", result) {
		// 	log.Debug().Str("uniquemember", m).Str("已存在", result).Msg("跳过")
		// 	continue
		// }
		modify := ldap.NewModifyRequest(fmt.Sprintf("cn=%s,ou=%s,%s", m, user.Group.Ou, l.dn), nil)
		attrVal := fmt.Sprintf("cn=%s,ou=%s,%s", user.Username, user.Ou, l.dn)
		modify.Add("uniquemember", []string{attrVal})
		err := l.conn.Modify(modify)
		if err != nil {
			log.Debug().Err(err).Msg("用户添加用户组失败")
			continue
		}
		log.Debug().Msgf("用户: %s 添加用户组: %s 成功", user.Username, m)
	}

	return nil
}

// DelUser
// 删除用户
func (l *ldapConn) DelUser(user *LdapUser) error {
	if user.Username == "" {
		// 中文转拼音
		user.Username = Pinyin(user.Zhusername)
	}
	// 判断用户是否存在
	if !l.SearchUser(user.Username) {
		log.Debug().Str("用户不存在", user.Username).Msg("")
		return errors.New("用户不存在")
	}
	distinguishedName := fmt.Sprintf("cn=%s,ou=%s,%s", user.Username, user.Ou, l.dn)
	ldaprow := ldap.NewDelRequest(distinguishedName, nil)
	log.Info().Str("删除用户", distinguishedName).Msg("")

	return l.conn.Del(ldaprow)
}

// ResetPasswordUser
// 重置密码用户
func (l *ldapConn) ResetPasswordUser(user *LdapUser) error {
	if user.Username == "" {
		// 中文转拼音
		user.Username = Pinyin(user.Zhusername)
	}
	// 判断用户是否存在
	if !l.SearchUser(user.Username) {
		log.Debug().Str("用户不存在", user.Username).Msg("")
		return errors.New("用户不存在")
	}
	if user.Password == "" {
		user.Password = RandomString(12)
		log.Debug().
			Str("用户名", user.Username).
			Str("随机生成密码为", user.Password).
			Msg("密码为空")
	}
	// 创建一个PasswordModifyRequest对象，用于修改密码
	distinguishedName := fmt.Sprintf("cn=%s,ou=%s,%s", user.Username, user.Ou, l.dn)
	passwordModifyRequest := ldap.NewPasswordModifyRequest(distinguishedName, "", user.Password)
	_, err := l.conn.PasswordModify(passwordModifyRequest)
	if err != nil {
		log.Debug().Err(err).
			Str("用户名", user.Username).
			Msg("重置密码失败")
	}
	log.Info().
		Str("用户名", user.Username).
		Str("密码", user.Password).
		Msg("重置密码用户")
	return err
}
