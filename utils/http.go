package utils

import (
	"github.com/valyala/fasthttp"
	"sync"
)

var reqPool = sync.Pool{New: func() interface{} {
	req := new(fasthttp.Request)
	req.Header.Set("User-Agent", "QQ/8.2.0.1296 CFNetwork/1126")
	req.Header.Set("Net-Type", "Wifi")
	return req
}}

func HttpGetBytes(url string) (data []byte, err error) {
	req := reqPool.Get().(*fasthttp.Request)
	resp := fasthttp.AcquireResponse()
	req.Header.SetMethod("GET")
	req.SetRequestURI(url)
	err = fasthttp.Do(req, resp)
	if err != nil {
		goto end
	}
	data = resp.Body()
end:
	reqPool.Put(req)
	fasthttp.ReleaseResponse(resp)
	return
}

func HttpPostBytes(url string, body []byte) (data []byte, err error) {
	req := reqPool.Get().(*fasthttp.Request)
	resp := fasthttp.AcquireResponse()
	req.Header.SetMethod("POST")
	req.SetRequestURI(url)
	req.SetBody(body)
	err = fasthttp.Do(req, resp)
	if err != nil {
		goto end
	}
	data = resp.Body()
end:
	req.ResetBody()
	reqPool.Put(req)
	fasthttp.ReleaseResponse(resp)
	return
}
