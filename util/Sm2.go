package util

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

type Sm2Crypt struct {
	// 公钥文件
	PublicFile string
	// 私钥文件
	PrivateFile string
	// 公钥字节
	PublicByte []byte
	// 私钥字节
	PrivateByte []byte
	// 证书路径
	CertPath string
	// 证书名称
	CertName string
}

// NewSm2Cert 用于创建一个新的SM2证书实例
// CertName: 证书名称，不包含文件扩展名
// CertPath: 证书文件存储的路径，必须以"/"结尾
// 返回值: 初始化后的Sm2Crypt指针
func NewSm2Cert(CertName, CertPath string) *Sm2Crypt {
	// 判断路径是否以/结尾，若不是则添加/
	if CertPath[len(CertPath)-1] != '/' {
		CertPath += "/"
	}
	// 初始化Sm2Crypt结构体，设置证书文件和私钥文件的路径
	crypt := Sm2Crypt{
		PublicFile:  CertPath + CertName + ".pem",
		PrivateFile: CertPath + CertName + ".key",
		CertPath:    CertPath,
		CertName:    CertName,
	}
	// 生成SM2密钥对
	crypt.GenerateSM2Key()
	// 读取PEM证书上下文
	crypt.readPemCxt()
	return &crypt
}

// readPemCxt 读取证书
// 该函数负责读取指定路径下的公钥和私钥文件，并将它们以字节切片的形式存储在Sm2Crypt结构体中。
// 参数:
//
//	s *Sm2Crypt: 指向Sm2Crypt结构体的指针，其中包含了证书的文件路径等信息。
//
// 返回值:
//
//	error: 如果读取过程中发生错误，将返回error；否则返回nil。
func (s *Sm2Crypt) readPemCxt() (err error) {

	// 读取公钥文件PEM格式内容
	s.PublicByte, err = os.ReadFile(s.PublicFile)
	if err != nil {
		return err
	}

	// 读取私钥文件PEM格式内容
	s.PrivateByte, err = os.ReadFile(s.PrivateFile)
	if err != nil {
		return err
	}

	return nil
}

// GenerateSM2Key 生成SM2密钥对，并将私钥和公钥保存到指定的文件中。
// 此函数首先检查私钥文件是否已经存在，如果存在，则不生成新的密钥对。
// 如果私钥文件不存在，则生成SM2私钥和公钥，并将它们分别保存到私钥文件和公钥文件中。
func (s *Sm2Crypt) GenerateSM2Key() {
	// 检查私钥文件是否已存在，若存在则不生成新的密钥对
	if _, err := os.Stat(s.PrivateFile); err == nil {
		//log.Println("文件已存在，跳过生成")
		return
	}

	// 生成SM2私钥和公钥
	priKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		//log.Println("秘钥产生失败：", err)
		os.Exit(1)
	}
	pubKey := &priKey.PublicKey

	// 生成私钥和公钥文件
	// 对私钥和公钥进行PEM编码，并分别写入到私钥文件和公钥文件中
	pemPrivKey, _ := x509.WritePrivateKeyToPem(priKey, nil)
	privateFile, _ := os.Create(s.PrivateFile)
	defer privateFile.Close()
	privateFile.Write(pemPrivKey)
	// 公钥使用PEM编码
	pemPublicKey, _ := x509.WritePublicKeyToPem(pubKey)
	publicFile, _ := os.Create(s.PublicFile)
	defer publicFile.Close()
	publicFile.Write(pemPublicKey)
}

// Encrypt 使用SM2加密算法对数据进行加密
// 参数:
//
//	data 需要加密的明文数据
//
// 返回值:
//
//	加密后的数据字符串和可能出现的错误
func (s *Sm2Crypt) Encrypt(data string) (string, error) {
	// 检查公钥证书字节是否提供
	if len(s.PublicByte) <= 0 {
		return "", errors.New("公钥证书字节长度为0")
	}

	// 从PEM格式的字节中读取公钥
	publicKeyFromPem, err := x509.ReadPublicKeyFromPem(s.PublicByte)
	if err != nil {
		// log.Println(err)
		return "", err
	}
	// 使用公钥进行ASN.1格式的加密
	ciphertxt, err := publicKeyFromPem.EncryptAsn1([]byte(data), rand.Reader)
	if err != nil {
		return "", err
	}
	// 将加密后的数据编码为Base64字符串
	return base64.StdEncoding.EncodeToString(ciphertxt), nil
}

// Decrypt 使用SM2算法解密给定的加密数据。
//
// 参数:
// data - 需要解密的数据，以Base64编码的字符串形式提供。
//
// 返回值:
// 解密后的文本字符串以及可能出现的错误。
// 如果私钥证书字节长度为0，或者在解密过程中遇到任何错误，将返回空字符串和相应的错误。
func (s *Sm2Crypt) Decrypt(data string) (string, error) {
	// 检查私钥证书字节是否为空
	if len(s.PrivateByte) <= 0 {
		return "", errors.New("私钥证书字节长度为0")
	}

	// 从PEM格式的字节中读取私钥
	privateKeyFromPem, err := x509.ReadPrivateKeyFromPem(s.PrivateByte, nil)
	if err != nil {
		return "", err
	}

	// 将Base64编码的加密数据解码为字节切片
	ciphertxt, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	// 使用私钥解密数据
	plaintxt, err := privateKeyFromPem.DecryptAsn1(ciphertxt)
	if err != nil {
		return "", err
	}

	return string(plaintxt), nil
}
