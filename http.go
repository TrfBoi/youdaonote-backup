package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

var (
	Timeout = 90
	client  = &http.Client{
		Timeout: time.Duration(Timeout) * time.Second,
	}

	RawCookies string
	Cookies    = make(map[string]*http.Cookie)
)

func init() {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client.Jar = cookieJar
}

// ParseCookies 将命令行参数设置的cookies解析
func ParseCookies() error {
	if len(RawCookies) == 0 {
		return fmt.Errorf("http RawCookies is empty")
	}
	RawCookies = strings.TrimSpace(RawCookies)
	cookieKVs := strings.Split(RawCookies, ";")
	for _, kv := range cookieKVs {
		kv = strings.TrimSpace(kv)
		kAndV := strings.Split(kv, "=")
		cookie := &http.Cookie{
			Name:  kAndV[0],
			Value: kAndV[1],
		}
		Cookies[cookie.Name] = cookie
	}
	return nil
}

func setCookies(rawURL string) {
	cookiesURL, _ := url.Parse(rawURL)
	cookies := make([]*http.Cookie, 0, len(Cookies))
	for _, v := range Cookies {
		cookies = append(cookies, v)
	}
	client.Jar.SetCookies(cookiesURL, cookies)
}

func Get(url string) ([]byte, error) {
	rsp, err := get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Close()
	return io.ReadAll(rsp)
}

func get(url string) (io.ReadCloser, error) {
	setCookies(url)
	rsp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		errData, err := io.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%v: %v", rsp.StatusCode, string(errData))
	}
	return rsp.Body, nil
}

// PostForm body以 application/x-www-form-urlencoded 格式设置的POST请求封装
func PostForm(rawURL string, body map[string]string) ([]byte, error) {
	setCookies(rawURL)
	form := url.Values(make(map[string][]string))
	if body != nil {
		for k, v := range body {
			form.Set(k, v)
		}
	}
	rsp, err := client.PostForm(rawURL, form)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		errData, err := io.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%v: %v", rsp.StatusCode, string(errData))
	}
	return io.ReadAll(rsp.Body)
}
