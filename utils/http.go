package utils

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var client = &http.Client{
	Transport: &http.Transport{
		ForceAttemptHTTP2:   true,
		MaxConnsPerHost:     0,
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 999,
	},
}

// HttpGetBytes 带 cookie 的 GET 请求
func HttpGetBytes(url, cookie string) ([]byte, error) {
	body, err := HTTPGetReadCloser(url, cookie)
	defer func() { _ = body.Close() }()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(body)
}

func HttpPostBytes(url string, data []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header["User-Agent"] = []string{"QQ/8.2.0.1296 CFNetwork/1126"}
	req.Header["Net-Type"] = []string{"Wifi"}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		buffer := bytes.NewBuffer(body)
		r, _ := gzip.NewReader(buffer)
		defer r.Close()
		unCom, err := ioutil.ReadAll(r)
		return unCom, err
	}
	return body, nil
}

func HttpPostBytesWithCookie(url string, data []byte, cookie string, contentType ...string) ([]byte, error) {
	t := "application/json"
	if len(contentType) > 0 {
		t = contentType[0]
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header["User-Agent"] = []string{"Dalvik/2.1.0 (Linux; U; Android 7.1.2; PCRT00 Build/N2G48H)"}
	req.Header["Content-Type"] = []string{t}
	if cookie != "" {
		req.Header["Cookie"] = []string{cookie}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		buffer := bytes.NewBuffer(body)
		r, _ := gzip.NewReader(buffer)
		defer r.Close()
		unCom, err := ioutil.ReadAll(r)
		return unCom, err
	}
	return body, nil
}

type gzipCloser struct {
	f io.Closer
	r *gzip.Reader
}

// NewGzipReadCloser 从 io.ReadCloser 创建 gunzip io.ReadCloser
func NewGzipReadCloser(reader io.ReadCloser) (io.ReadCloser, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	return &gzipCloser{
		f: reader,
		r: gzipReader,
	}, nil
}

// Read impls io.Reader
func (g *gzipCloser) Read(p []byte) (n int, err error) {
	return g.r.Read(p)
}

// Close impls io.Closer
func (g *gzipCloser) Close() error {
	_ = g.f.Close()
	return g.r.Close()
}

// HTTPGetReadCloser 从 Http url 获取 io.ReadCloser
func HTTPGetReadCloser(url string, cookie string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header["User-Agent"] = []string{"QQ/8.2.0.1296 CFNetwork/1126"}
	req.Header["Net-Type"] = []string{"Wifi"}
	if cookie != "" {
		req.Header["Cookie"] = []string{cookie}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		return NewGzipReadCloser(resp.Body)
	}
	return resp.Body, err
}
