package mail

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DKIMOptions 存储DKIM签名选项
type DKIMOptions struct {
	Domain            string    // 邮件域名
	Selector          string    // DKIM选择器名称
	PrivateKeyPath    string    // 私钥文件路径
	HeadersToSign     []string  // 要签名的邮件头
	SignatureExpireIn int64     // 签名过期时间（秒）
	BodyLength        int       // 签名的邮件正文长度，0表示全部
}

// DKIMSigner 处理DKIM签名
type DKIMSigner struct {
	options     DKIMOptions
	privateKey  *rsa.PrivateKey
	initialized bool
}

// NewDKIMSigner 创建一个新的DKIM签名器
func NewDKIMSigner(options DKIMOptions) (*DKIMSigner, error) {
	signer := &DKIMSigner{
		options: options,
	}
	
	// 设置默认值
	if len(options.HeadersToSign) == 0 {
		signer.options.HeadersToSign = []string{
			"From", "To", "Subject", "Date", "MIME-Version",
			"Content-Type", "Content-Transfer-Encoding", "Message-ID",
		}
	}
	
	// 加载私钥
	if err := signer.loadPrivateKey(); err != nil {
		return nil, err
	}
	
	signer.initialized = true
	return signer, nil
}

// 加载私钥
func (s *DKIMSigner) loadPrivateKey() error {
	if s.options.PrivateKeyPath == "" {
		return fmt.Errorf("未指定DKIM私钥路径")
	}
	
	// 读取私钥文件
	privateKeyData, err := ioutil.ReadFile(s.options.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("读取DKIM私钥文件失败: %v", err)
	}
	
	// 解码PEM格式
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return fmt.Errorf("无法解码PEM格式的私钥")
	}
	
	// 解析RSA私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析RSA私钥失败: %v", err)
	}
	
	s.privateKey = privateKey
	return nil
}

// SignMessage 对邮件内容进行DKIM签名
func (s *DKIMSigner) SignMessage(messageData []byte) ([]byte, error) {
	if !s.initialized || s.privateKey == nil {
		return nil, fmt.Errorf("DKIM签名器未正确初始化")
	}
	
	// 解析邮件分割头部和正文
	headerEnd := bytes.Index(messageData, []byte("\r\n\r\n"))
	if headerEnd == -1 {
		headerEnd = bytes.Index(messageData, []byte("\n\n"))
		if headerEnd == -1 {
			return nil, fmt.Errorf("无法解析邮件格式")
		}
	}
	
	headers := messageData[:headerEnd]
	body := messageData[headerEnd:]
	
	// 解析头部为键值对
	headerMap := parseHeaders(headers)
	
	// 计算正文的哈希值
	bodyHash := computeBodyHash(body)
	
	// 生成签名头部
	signatureHeader := s.createSignatureHeader(headerMap, bodyHash)
	
	// 计算签名
	signature, err := s.computeSignature(signatureHeader)
	if err != nil {
		return nil, fmt.Errorf("计算DKIM签名失败: %v", err)
	}
	
	// 将签名添加到签名头部
	dkimHeader := fmt.Sprintf("%s b=%s", signatureHeader, signature)
	
	// 将DKIM签名头部添加到邮件头部
	result := append([]byte("DKIM-Signature: "+dkimHeader+"\r\n"), messageData...)
	
	log.Printf("成功添加DKIM签名: 域名=%s, 选择器=%s", s.options.Domain, s.options.Selector)
	return result, nil
}

// parseHeaders 解析邮件头部
func parseHeaders(headers []byte) map[string]string {
	headerMap := make(map[string]string)
	lines := bytes.Split(headers, []byte("\n"))
	
	var currentKey, currentValue string
	
	for _, line := range lines {
		// 删除回车符
		line = bytes.TrimSuffix(line, []byte("\r"))
		
		// 如果行以空白字符开始，它是上一行的继续
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			currentValue += string(bytes.TrimSpace(line))
			headerMap[currentKey] = currentValue
			continue
		}
		
		// 否则，这是一个新的头部
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			continue
		}
		
		currentKey = string(bytes.TrimSpace(parts[0]))
		currentValue = string(bytes.TrimSpace(parts[1]))
		headerMap[currentKey] = currentValue
	}
	
	return headerMap
}

// computeBodyHash 计算邮件正文的哈希值
func computeBodyHash(body []byte) string {
	// 标准化行尾（CRLF）
	body = bytes.ReplaceAll(body, []byte("\r\n"), []byte("\n"))
	body = bytes.ReplaceAll(body, []byte("\n"), []byte("\r\n"))
	
	// 计算SHA-256哈希
	h := sha256.New()
	h.Write(body)
	
	// 返回Base64编码的哈希值
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// createSignatureHeader 创建DKIM签名头部（不包括签名本身）
func (s *DKIMSigner) createSignatureHeader(headers map[string]string, bodyHash string) string {
	// 当前时间戳
	now := time.Now().Unix()
	// 生成随机标识符
	randID := fmt.Sprintf("%d.%d", now, rand.Int())
	
	// 构建头部字段列表
	var headerFields []string
	for _, header := range s.options.HeadersToSign {
		if _, exists := headers[header]; exists {
			headerFields = append(headerFields, strings.ToLower(header))
		}
	}
	
	// 构建签名头部
	signature := fmt.Sprintf(
		"v=1; a=rsa-sha256; c=relaxed/relaxed; d=%s; s=%s; t=%d; i=@%s; h=%s; bh=%s;",
		s.options.Domain,
		s.options.Selector,
		now,
		s.options.Domain,
		strings.Join(headerFields, ":"),
		bodyHash,
	)
	
	// 添加过期时间（如果设置）
	if s.options.SignatureExpireIn > 0 {
		expiry := now + s.options.SignatureExpireIn
		signature += fmt.Sprintf(" x=%d;", expiry)
	}
	
	return signature
}

// computeSignature 计算DKIM签名
func (s *DKIMSigner) computeSignature(signatureHeader string) (string, error) {
	// 准备数据进行签名
	h := sha256.New()
	h.Write([]byte(signatureHeader))
	
	// 使用私钥对哈希值签名
	signature, err := rsa.SignPKCS1v15(nil, s.privateKey, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return "", err
	}
	
	// 返回Base64编码的签名
	return base64.StdEncoding.EncodeToString(signature), nil
}

// GenerateDKIMKey 生成DKIM密钥对
func GenerateDKIMKey(domain, selector string, keyLength int) error {
	// 创建密钥目录
	keyDir := filepath.Join("keys", domain)
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("创建密钥目录失败: %v", err)
	}
	
	// 设置默认的密钥长度
	if keyLength <= 0 {
		keyLength = 2048 // RSA 2048位是一个良好的默认值
	}
	
	// 构建文件路径
	privateKeyPath := filepath.Join(keyDir, fmt.Sprintf("%s.private", selector))
	publicKeyPath := filepath.Join(keyDir, fmt.Sprintf("%s.public", selector))
	dnsRecordPath := filepath.Join(keyDir, fmt.Sprintf("%s.txt", selector))
	
	// 检查是否已存在
	if _, err := os.Stat(privateKeyPath); err == nil {
		return fmt.Errorf("密钥已存在: %s", privateKeyPath)
	}
	
	// 生成私钥 (这里执行系统命令，需要有openssl)
	log.Printf("生成DKIM私钥: %s", privateKeyPath)
	cmd := exec.Command("openssl", "genrsa", "-out", privateKeyPath, fmt.Sprintf("%d", keyLength))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("生成私钥失败: %v", err)
	}
	
	// 提取公钥
	log.Printf("提取DKIM公钥: %s", publicKeyPath)
	cmd = exec.Command("openssl", "rsa", "-in", privateKeyPath, "-pubout", "-outform", "PEM", "-out", publicKeyPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("提取公钥失败: %v", err)
	}
	
	// A workaround for "inappropriate ioctl for device" error when running openssl command
	if fileInfo, err := os.Stat(publicKeyPath); err != nil || fileInfo.Size() == 0 {
		// Generate RSA key pair directly from Go
		key, err := rsa.GenerateKey(rand.Reader, keyLength)
		if err != nil {
			return fmt.Errorf("生成RSA密钥对失败: %v", err)
		}
		
		// Save private key
		privateKeyPEM := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}
		privateKeyData := pem.EncodeToMemory(privateKeyPEM)
		if err := os.WriteFile(privateKeyPath, privateKeyData, 0600); err != nil {
			return fmt.Errorf("保存私钥失败: %v", err)
		}
		
		// Save public key
		publicKeyPEM := &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
		}
		publicKeyData := pem.EncodeToMemory(publicKeyPEM)
		if err := os.WriteFile(publicKeyPath, publicKeyData, 0644); err != nil {
			return fmt.Errorf("保存公钥失败: %v", err)
		}
	}
	
	// 读取公钥内容并转换为DNS记录格式
	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("读取公钥文件失败: %v", err)
	}
	
	// 提取公钥内容 (去除头尾和换行)
	publicKeyContent := string(publicKeyBytes)
	publicKeyContent = strings.ReplaceAll(publicKeyContent, "-----BEGIN PUBLIC KEY-----", "")
	publicKeyContent = strings.ReplaceAll(publicKeyContent, "-----END PUBLIC KEY-----", "")
	publicKeyContent = strings.ReplaceAll(publicKeyContent, "\n", "")
	publicKeyContent = strings.TrimSpace(publicKeyContent)
	
	// 保存为DNS记录格式
	dnsRecord := fmt.Sprintf("v=DKIM1; k=rsa; p=%s", publicKeyContent)
	if err := ioutil.WriteFile(dnsRecordPath, []byte(dnsRecord), 0644); err != nil {
		return fmt.Errorf("保存DNS记录失败: %v", err)
	}
	
	log.Printf("DKIM密钥生成完成!")
	log.Printf("私钥文件: %s", privateKeyPath)
	log.Printf("公钥文件: %s", publicKeyPath)
	log.Printf("DNS记录: %s", dnsRecordPath)
	log.Printf("请添加以下DNS记录:")
	log.Printf("%s._domainkey.%s. IN TXT \"%s\"", selector, domain, dnsRecord)
	
	return nil
}