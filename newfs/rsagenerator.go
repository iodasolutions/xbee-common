package newfs

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/util"
	"golang.org/x/crypto/ssh"
	"io"
	"sync"
)

//key.pem
//key.pub
//ca.pem

//server.crt
//server.key

var lock = &sync.Mutex{}

type RsaGenerator struct {
	sshFolder Folder
}

func NewRsaGen(folder Folder) *RsaGenerator {
	if folder == "" {
		folder = XbeeIntern().SSHFolder()
	}
	return &RsaGenerator{
		sshFolder: folder,
	}
}

func (rg *RsaGenerator) Root() Folder {
	return rg.sshFolder
}

func (rg *RsaGenerator) RootKeyPEM() File {
	return rg.sshFolder.ChildFile("key.pem")
}
func (rg *RsaGenerator) RootAuthorizedKey() File {
	return rg.sshFolder.ChildFile("key.pub")
}
func (rg *RsaGenerator) CAFile() File {
	return rg.sshFolder.ChildFile("ca.pem")
}
func (rg *RsaGenerator) ServerCert() File {
	return rg.sshFolder.ChildFile("server.crt")
}
func (rg *RsaGenerator) ServerKey() File {
	return rg.sshFolder.ChildFile("server.key")
}

func (rg *RsaGenerator) HasServerCertAndKey() bool {
	return rg.CAFile().Exists() && rg.ServerCert().Exists() && rg.ServerKey().Exists()
}
func (rg *RsaGenerator) HasRootKeys() bool {
	return rg.CAFile().Exists() && rg.RootKeyPEM().Exists() && rg.RootAuthorizedKey().Exists()
}
func (rg *RsaGenerator) CA() *x509.Certificate {
	p, _ := pem.Decode(rg.CAFile().ContentBytes())
	ca, err := x509.ParseCertificate([]byte(p.Bytes))
	if err != nil {
		panic(err)
	}
	return ca
}
func (rg *RsaGenerator) RootKey() *rsa.PrivateKey {
	p, _ := pem.Decode(rg.RootKeyPEM().ContentBytes())
	rootKey, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		panic(err)
	}
	return rootKey
}
func (rg *RsaGenerator) EnsureRootKeysExist(ctx context.Context) {
	lock.Lock()
	defer lock.Unlock()
	if !rg.HasRootKeys() {
		log2.Infof("Generate Xbee Key...")
		rg.createAndPersistRootCertificate()
	}
}
func (rg *RsaGenerator) createAndPersistRootCertificate() {
	rg.sshFolder.EnsureEmpty()
	rg.sshFolder.ChMod(0700)
	ca, caPrivKey := util.NewRootCA()
	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		panic(err)
	}

	// ca
	caPEM := rg.CAFile().OpenFileForCreation()
	defer caPEM.Close()
	if err := pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		panic(err)
	}
	rg.CAFile().ChMod(0644)

	// key
	caPrivKeyPEM := rg.RootKeyPEM().OpenFileForCreation()
	defer caPrivKeyPEM.Close()
	if err := pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	}); err != nil {
		panic(err)
	}
	rg.RootKeyPEM().ChMod(0600)

	pub, err := ssh.NewPublicKey(&caPrivKey.PublicKey)
	if err != nil {
		panic(fmt.Errorf("\nCannot return public key from private key : %v\n", err))
	}
	publicKeyWriter := rg.RootAuthorizedKey().OpenFileForCreation()
	defer publicKeyWriter.Close()
	buf := bytes.NewBuffer(ssh.MarshalAuthorizedKey(pub))
	if _, err := io.Copy(publicKeyWriter, buf); err != nil {
		panic(err)
	}
	rg.RootAuthorizedKey().ChMod(0644)

}

func (rg *RsaGenerator) NewServerCertificate() (certPEM []byte, privateKeyPEM []byte) {

	rootKey := rg.RootKey()
	ca := rg.CA()

	certPEM, privateKeyPEM = util.NewServerCertificate(ca, rootKey)
	return
	//serverTLSConf = &tls.Config{
	//	Certificates: []tls.Certificate{serverCert},
	//}

	//certpool := x509.NewCertPool()
	//certpool.AppendCertsFromPEM(caPEM.Bytes())
	//clientTLSConf = &tls.Config{
	//	RootCAs: certpool,
	//}

}

func (rg *RsaGenerator) ServerCertificate() *tls.Certificate {
	certPEM := rg.ServerCert().ContentBytes()
	privateKeyPEM := rg.ServerKey().ContentBytes()

	serverCert, err := tls.X509KeyPair(certPEM, privateKeyPEM)
	if err != nil {
		panic(err)
	}
	return &serverCert
}

func (rg *RsaGenerator) RootCertificate() *tls.Certificate {
	certPEM := rg.CAFile().ContentBytes()
	privateKeyPEM := rg.RootKeyPEM().ContentBytes()

	serverCert, err := tls.X509KeyPair(certPEM, privateKeyPEM)
	if err != nil {
		panic(err)
	}
	return &serverCert
}

func (rg *RsaGenerator) CertPool() *x509.CertPool {
	certPool := x509.NewCertPool()

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(rg.CAFile().ContentBytes()); !ok {
		panic(errors.New("failed to append client certs"))
	}
	return certPool
}
