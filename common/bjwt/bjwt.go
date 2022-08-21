package bjwt

//package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"io"
	"strconv"
	"time"

	"fmt"
	//jwt "github.com/dgrijalva/jwt-go"
	"strings"
)

const (
	TimeFormat    = "2006-01-02 15:04:05"
	SecretKey     = "Djql@#al1E4$Bl"
	AesDefualtKey = "GKgKRbvX1Bu&23Nk"
	ExpireTime    = 24 * 12 * 3600
)

type Header struct {
	Alg  string `json:"alg"` //校验算法
	Type string `json:"type"`
	ScrV string `json:"scr_v"`
}

type Claims struct {
	//Username string `json:"username"`
	//Uid string `json:"uid"`
	Id        string `json:"id"`
	Usertype  int64  `json:"usertype"` //用户类型
	Expr      int64  `json:"expr"`     //超时时间
	TimeStamp int64  `json:"timestamp"`
}

func CreatHeader() Header {
	return Header{
		Alg:  "HS256",
		Type: "BJWT",
		ScrV: strconv.FormatInt(time.Now().Unix()/30, 10),
	}
}

func CreateClaims(id string, ut int64) Claims {
	now := time.Now().Unix()
	return Claims{
		//Username:  un,
		//Uid:       uid,
		Id:        id,
		Usertype:  ut,
		Expr:      now + ExpireTime,
		TimeStamp: now,
	}
}

func GetId(c Claims) string {
	return c.Id
}

func RefreshToken(h Header, c Claims) string {
	now := time.Now().Unix()
	c.TimeStamp = now
	c.Expr = now + ExpireTime

	return CreateToken(h, c)
}

func IsAnonymousClaims(cliams Claims) bool {
	if cliams.Id == "anonymous" {
		return true
	} else {
		return false
	}
}

func CreateAnonymousClaims() Claims {
	return Claims{
		//Username:  "anonymous",
		//Uid:       "anonymous",
		Id:        "anonymous",
		Expr:      time.Now().Unix() + 120,
		TimeStamp: time.Now().Unix(),
	}
}

func HmacHs256(message string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	io.WriteString(h, message)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func CheckToken(token string) (int, Header, Claims, string) {
	//check token
	cur := time.Now().Unix()
	hs, cs, ss, err := ExtractToken(token)

	//
	// For debug , can use forever !!!!
	//
	if token == "eyJhbGciOiJIUzI1NiIsInR5cGUiOiJCSldUIiwic2NyX3YiOiI1MzAwNDQyMSJ9.qA1eo7TdPjW8dtECsrUyqAun3euXLgjL8HLkSwwlEv7kTe2KC271ajw8r+D310qAx2DtKBAJYQ9QctZbuMxyXGZSNQMOSxj7eQ9knRHA/ZLHxwAeMBmZso3eP3rmusTPFI43sZPEU5Q1qvDjsiGYug==.0995dc94137393cfee1549cd522eeaa26d2dfbe538210907bcc9f8dc8696adfc" {
		return 0, hs, cs, ss //未超时token
	}
	//
	//

	if err != nil { //无效token
		return 1, hs, cs, ss
	}
	if cur >= cs.Expr && IsAnonymousClaims(cs) {
		return 100, hs, cs, ss
	}
	if cur >= cs.Expr { //超时
		return 2, hs, cs, ss
	}
	if IsAnonymousClaims(cs) {
		return 3, hs, cs, ss
	}
	return 0, hs, cs, ss //未超时token
}
func ExtractToken(token string) (Header, Claims, string, error) {
	var h *Header = &Header{}
	var c *Claims = &Claims{}
	var s string = ""

	as := strings.Split(token, ".")
	if len(as) != 3 {
		return *h, *c, s, errors.New("illegal token")
	}

	//校验签名
	hc := as[0] + "." + as[1]
	if as[2] != HmacHs256(hc, SecretKey) {
		return *h, *c, s, errors.New("illegal token")
	}

	//base64解码
	h_s, err1 := base64.StdEncoding.DecodeString(as[0])
	if err1 != nil {
		return *h, *c, s, errors.New("illegal token")
	}
	c_aes, err2 := base64.StdEncoding.DecodeString(as[1])
	if err2 != nil {
		return *h, *c, s, errors.New("illegal token")
	}

	//aes解密body
	c_s, err5 := AesDecrypt(c_aes, []byte(AesDefualtKey))
	//result, err := AesEncrypt([]byte("我的名字叫吴锦锋哈哈哈"), key)
	if err5 != nil {
		return *h, *c, s, errors.New("illegal token")
	}

	//解析json
	err3 := json.Unmarshal([]byte(h_s), h)
	if err3 != nil {
		return *h, *c, s, errors.New("illegal token")
	}
	err4 := json.Unmarshal([]byte(c_s), c)
	if err4 != nil {
		return *h, *c, s, errors.New("illegal token")
	}

	s = as[2]

	return *h, *c, s, nil
}

func CreateToken(h Header, c Claims) string {
	//Json 格式化
	h_b, _ := json.Marshal(h)
	c_b, _ := json.Marshal(c)

	c_aes, _ := AesEncrypt(c_b, []byte(AesDefualtKey))

	//Base64化
	h_s := base64.StdEncoding.EncodeToString(h_b)
	c_s := base64.StdEncoding.EncodeToString([]byte(c_aes))

	hc := h_s + "." + c_s
	sign := HmacHs256(hc, SecretKey)

	return hc + "." + sign
}

func main() {
	var h *Header = &Header{}
	var c *Claims = &Claims{}
	//var s string = ""

	h.Alg = "HS256"
	h.ScrV = "673350670423"
	h.Type = "BJWT"

	c.Expr = time.Now().Unix() + ExpireTime
	c.TimeStamp = time.Now().Unix()
	c.Id = "我的名字叫吴锦锋"
	//c.Uid = "wujinfeng"
	//c.Username = "我的名字叫吴锦锋"

	token := CreateToken(*h, *c)
	fmt.Println(token)

	_, c1, _, _ := ExtractToken(token)
	fmt.Println("解密到结构体后打印ID:")
	fmt.Println(c1.Id)
}
