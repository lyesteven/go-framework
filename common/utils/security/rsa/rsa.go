package rsa

import (
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "errors"
    "crypto/sha256"
    "encoding/base64"
)

func SignRSA2(src, key []byte) (string, error) {
    sig := ""

    var h = sha256.New()
    h.Write(src)
    var hashed = h.Sum(nil)

    var err error
    var block *pem.Block
    block, _ = pem.Decode(key)
    if block == nil {
        return sig, errors.New("private key error")
    }

    var pri *rsa.PrivateKey
    pri, err = x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        return sig, err
    }
    sigByte, err := rsa.SignPKCS1v15(rand.Reader, pri, crypto.SHA256, hashed)
    if err != nil {
        return sig, err
    }

    sig = base64.StdEncoding.EncodeToString(sigByte)
    return sig, nil
}

func VerifyRSA2(src, key []byte, sig string) error {
    sigByte, err := base64.StdEncoding.DecodeString(sig)
    if err != nil {
        return err
    }

    var h = sha256.New()
    h.Write(src)
    var hashed = h.Sum(nil)

    var block *pem.Block
    block, _ = pem.Decode(key)
    if block == nil {
        return errors.New("public key error")
    }

    var pubInterface interface{}
    pubInterface, err = x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return err
    }
    var pub = pubInterface.(*rsa.PublicKey)

    return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed, sigByte)
}