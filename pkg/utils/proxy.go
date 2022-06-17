package utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

type ISocks5Proxy interface {
	Dial(network, addr string) (c net.Conn, err error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type Socks5Proxy struct {
	dial        func(network, addr string) (c net.Conn, err error)
	dialContext func(ctx context.Context, network, address string) (net.Conn, error)
}

func (s *Socks5Proxy) Dial(network, addr string) (c net.Conn, err error) {
	return s.dial(network, addr)
}

func (s *Socks5Proxy) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return s.dialContext(ctx, network, address)
}

/*创建一个socks5代理
  address 代理地址. 支持socks5, socks5h. 示例: socks5://127.0.0.1:1080 socks5://user:pwd@127.0.0.1:1080
*/
func NewSocks5Proxy(address string) (ISocks5Proxy, error) {
	// 解析地址
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("address无法解析: %v", err)
	}

	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "socks5", "socks5h":
		var auth *proxy.Auth
		if u.User != nil {
			pwd, ok := u.User.Password()
			auth = &proxy.Auth{User: u.User.Username()}
			if ok {
				auth.Password = pwd
			}
		}

		dialer, err := proxy.SOCKS5("tcp", u.Host, auth, nil)
		if err != nil {
			return nil, fmt.Errorf("dialer生成失败: %v", err)
		}

		var dialCtx func(ctx context.Context, network, address string) (net.Conn, error)
		if d, ok := dialer.(proxy.ContextDialer); ok {
			dialCtx = d.DialContext
		} else {
			dialCtx = func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.Dial(network, address)
			}
		}

		sp := &Socks5Proxy{
			dial:        dialer.Dial,
			dialContext: dialCtx,
		}
		return sp, nil
	}
	return nil, fmt.Errorf("address的scheme不支持: %s")
}

type IHttpProxy interface {
	SetProxy(transport *http.Transport)
}

type HttpProxy struct {
	p  func(request *http.Request) (*url.URL, error)
	s5 ISocks5Proxy
}

func (h *HttpProxy) SetProxy(transport *http.Transport) {
	if h.s5 != nil {
		transport.DialContext = h.s5.DialContext
		return
	}

	transport.Proxy = h.p
}

/*创建一个http代理
  address 代理地址. 支持 http, https, socks5, socks5h. 示例: https://127.0.0.1:1080 https://user:pwd@127.0.0.1:1080
*/
func NewHttpProxy(address string) (IHttpProxy, error) {
	// 解析地址
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("address无法解析: %v", err)
	}

	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "http", "https":
		p := func(request *http.Request) (*url.URL, error) {
			return u, nil
		}
		return &HttpProxy{p: p}, nil
	case "socks5", "socks5h":
		s5, err := NewSocks5Proxy(address)
		if err != nil {
			return nil, err
		}
		return &HttpProxy{s5: s5}, nil
	}
	return nil, fmt.Errorf("address的scheme不支持: %s")
}
