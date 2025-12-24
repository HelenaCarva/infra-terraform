package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type AwsClient struct {
	IAM    *iam.IAM
	STS    *sts.STS
	SSM    *ssm.SSM
	Session *session.Session
}

func NewAwsClient() (*AwsClient, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))}, nil)
	if err!= nil {
		return nil, err
	}

	iamClient := iam.New(sess)
	stsClient := sts.New(sess)
	ssmClient := ssm.New(sess)

	return &AwsClient{
		IAM:    iamClient,
		STS:    stsClient,
		SSM:    ssmClient,
		Session: sess,
	}, nil
}

func GenerateKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err!= nil {
		return "", "", err
	}

	privatePEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	privatePemData := pem.EncodeToMemory(privatePEM)

	publicPEM := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKIXPublicKey(&privateKey.PublicKey)}
	publicPemData := pem.EncodeToMemory(publicPEM)

	return string(privatePemData), string(publicPemData), nil
}

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err!= nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok &&!ipnet.IP.IsLoopback() {
			if ipnet.IP.To4()!= nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", nil
}

func GetRandomString(length int) (string, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6
		letterIdxMask = 1 << letterIdxBits - 1
	)
	letterIdx := 0
	b := make([]byte, length)
	for i := range b {
		if letterIdx == 0 {
			letterIdx = int(rand.Int63() & letterIdxMask)
		}
		b[i] = letterBytes[letterIdx]
		letterIdx = (letterIdx + 1) & (letterIdxMask - 1)
	}
	return string(b), nil
}

func GetCertSANs(dnsNames []string, ips []string) []string {
	sans := make([]string, 0, len(dnsNames)+len(ips))
	for _, dnsName := range dnsNames {
		sans = append(sans, dnsName)
	}
	for _, ip := range ips {
		sans = append(sans, ip)
	}
	return sans
}

func GetCertSANsFromHost(hostname string) []string {
	return []string{hostname}
}

func GetCertSANsFromEnvironmentVariables() []string {
	return []string{os.Getenv("HOSTNAME")}
}

func GetCertSANsFromConfig() []string {
	return []string{os.Getenv("HOSTNAME")}
}

func GenerateCertSANs() ([]string, error) {
	return GetCertSANsFromConfig()
}

func GenerateCert(hostname string, notBefore, notAfter time.Time, sans []string) (*x509.Certificate, error) {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Example Organization"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		DNSNames:  sans,
	}

	return template, nil
}

func GenerateCertFromConfig() (*x509.Certificate, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	return GenerateCert(os.Getenv("HOSTNAME"), notBefore, notAfter, GenerateCertSANs())
}