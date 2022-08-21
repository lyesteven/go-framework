package des

import (
    "bytes"
    "crypto/des"
    "crypto/cipher"
    "unicode/utf16"
    "encoding/binary"
)

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
    padding := blockSize - len(ciphertext)%blockSize
    padtext := bytes.Repeat([]byte{byte(padding)}, padding)
    return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
    length := len(origData)
    unpadding := int(origData[length-1])
    return origData[:(length - unpadding)]
}

func ConvertUTF16ToLittleEndianBytes(s string) []byte {
    u := utf16.Encode([]rune(s))
    b := make([]byte, 2*len(u))
    for index, value := range u {
        binary.LittleEndian.PutUint16(b[index*2:], value)
    }
    return b
}

func DesEncrypt(origData, key []byte) ([]byte, error) {
    block, err := des.NewCipher(key)
    if err != nil {
        return nil, err
    }

    origData = PKCS5Padding(origData, block.BlockSize())
    blockMode := cipher.NewCBCEncrypter(block, key)
    crypted := make([]byte, len(origData))
    blockMode.CryptBlocks(crypted, origData)
    return crypted, nil
}

func DesDecrypt(crypted, key []byte) ([]byte, error) {
    block, err := des.NewCipher(key)
    if err != nil {
        return nil, err
    }
    blockMode := cipher.NewCBCDecrypter(block, key)
    origData := make([]byte, len(crypted))
    blockMode.CryptBlocks(origData, crypted)
    origData = PKCS5UnPadding(origData)
    return origData, nil
}

