package main

import (
	"github.com/valyala/fasthttp"
	"time"
	"fmt"

	"github.com/lyesteven/go-framework/common/bjwt"
)

func getTempToken(c *fasthttp.RequestCtx)  {
	header := bjwt.CreatHeader()
	claims := bjwt.CreateAnonymousClaims()

	Temp_token := bjwt.CreateToken(header, claims)
	expire := time.Now().Unix()+30

	retJson := fmt.Sprintf("{\"code\":0,\"message\":\"OK\",\"data\":{\"token\":\"%s\",\"expire\":%d},\"attachtype\":\"nouse\",\"attach\":\"\"}", Temp_token,expire)
	//retJson := fmt.Sprintf("{\"token\":\"%s\",\"expire\":%d}",Temp_token,expire)
	c.SuccessString("application/json",retJson )
}

func testTest(c *fasthttp.RequestCtx) {
	//c.JSON(http.StatusOK, time.Now().Unix())
	retJson := fmt.Sprintf("{\"code\":0,\"message\":\"%s\",\"data\":\"{\"Data\":\"%d\"}\",\"attachtype\":\"nouse\",\"attach\":\"\"}", time.Now().Unix())
	c.SuccessString("application/json",retJson )
}


