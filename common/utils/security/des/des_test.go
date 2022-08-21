package des

import (
    "testing"
    "encoding/base64"
)

func TestDesEncrypt(t *testing.T) {
    desKey := "12345678"
    t.Log(desKey)

    originData := `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC525iaO1g+tXx2RNqPksRGUjwK
1Pd3cRTsIxt/D/FMx0pli26GfwMThWQDn7tJ1MvwBqNw+MuYPwQftEsHAkLXF0Kn
8dBjT0PpflXbJwLgNM4W9T8LNN0a5+EEm83EkVEGcOWqSjMNSUNxIIdTUNdihcWF
v+zHBIvaYWBxX6h58QIDAQAB
-----END PUBLIC KEY-----`
    t.Log(originData)
    desKeyBytes := []byte(desKey)[:8]
    decryptData, err := DesEncrypt([]byte(originData), desKeyBytes)
    if err != nil {
        t.Fatal(err)
    }
    t.Log(base64.StdEncoding.EncodeToString(decryptData))

    ret, err := DesDecrypt(decryptData, desKeyBytes)
    if err != nil {
        t.Fatal(err)
    }
    t.Log(string(ret))

    if originData != string(ret) {
        t.Fatal("not equal")
    }
}
