package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// PostHttps 使用证书发起Https Post请求
func PostHttps(url string, data interface{}, contentType, certFile, keyFile string) ([]byte, error) {
	// 创建证书池及各类对象
	var client *http.Client
	var resp *http.Response
	var body []byte
	var err error

	jsonStr, _ := json.Marshal(data) // json
	var cliCrt tls.Certificate       // 具体的证书加载对象
	cliCrt, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// 把上面的准备内容传入 client
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cliCrt},
				InsecureSkipVerify: true,
			},
		},
	}

	// Get 请求
	resp, err = client.Post(url, contentType, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer client.CloseIdleConnections()
	return body, nil
}
