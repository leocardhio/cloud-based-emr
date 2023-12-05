package utils

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"service-pharmacy/config"
	"service-pharmacy/logger"
)

func GenerateSignature(data string) string {
	rsaPrivate, err := GetPrivateKey()
	if err != nil {
		logger.LogPanic.Panicf("Failed to get Signature Key")
	}

	hashed := sha256.Sum256([]byte(data))

	signature, err := rsa.SignPKCS1v15(nil, rsaPrivate, crypto.SHA256, hashed[:])
	if err != nil {
		logger.LogPanic.Panicf("Failed to signed document")
	}

	return base64.StdEncoding.EncodeToString(signature)

}

func GetPrivateKey() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(config.RSAPrivateKey))
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch key.(type) {
	case *rsa.PrivateKey:
		if pkey, ok := key.(*rsa.PrivateKey); ok {
			return pkey, nil
		}
	default:
		panic("unknown key")
	}

	panic("fail to get key")
}

func VerifySignature(docData string, sign string) (bool, error) {
	block, _ := pem.Decode([]byte(config.RSAPublicKey))
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("key not valid")
	}

	isRsaPublic := false
	switch key.(type) {
	case *rsa.PublicKey:
		isRsaPublic = true
	default:
		panic("unknown key")
	}

	if !isRsaPublic {
		panic("public not RSA")
	}

	pubkey := key.(*rsa.PublicKey)
	signatureByte, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		panic("signature undecodable")
	}

	validateByte := sha256.Sum256([]byte(docData))

	err = rsa.VerifyPKCS1v15(pubkey, crypto.SHA256, validateByte[:], signatureByte)
	if err != nil {
		return false, err
	}

	return true, nil
}
