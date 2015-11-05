package socketio

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"
	"log"
	"encoding/json"
)

// Session holds the configuration variables received from the socket.io
// server.
type Session struct {
	ID                 string
	HeartbeatTimeout   time.Duration
	ConnectionTimeout  time.Duration
	SupportedProtocols []string
}

// NewSession receives the configuraiton variables from the socket.io
// server.
func NewSession(url string) (*Session, error) {
	urlParser, err := newURLParser(url)
	if err != nil {
		return nil, err
	}
	response, err := http.Get(urlParser.handshake())

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body.Close()

	// 如果结果正确，则返回数据如下
	// [	?0{"sid":"E4Z4PNc4LZ30as06AAAa","upgrades":["websocket"],"pingInterval":25000,"pingTimeout":60000}]
	// 不包含中括号，中括号是用来标明字符串前面的空格符号
	// @TODO 不明确返回数据头部空格是如何产生的，后续反解析JSON数据的过程中需要去除此部分数据
	type HandShake struct {
		Sid  string
		Upgrades []string
		PingInterval int
		PingTimeout int
	}

	var shake HandShake
	err = json.Unmarshal(body[5:], &shake)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("Handshaking failed!")
	}

	/*
	sessionVars := strings.Split(string(body), ":")
	if len(sessionVars) != 4 {
		return nil, errors.New("Session variables is not 4")
	}
	*/

	id := shake.Sid

	heartbeatTimeoutSec := shake.PingInterval/1000
	connectionTimeoutSec := shake.PingTimeout/1000

	heartbeatTimeout := time.Duration(heartbeatTimeoutSec) * time.Second
	connectionTimeout := time.Duration(connectionTimeoutSec) * time.Second

	supportedProtocols := shake.Upgrades //strings.Split(string(body), ",")

	return &Session{id, heartbeatTimeout, connectionTimeout, supportedProtocols}, nil
}

// SupportProtocol checks if the given protocol is supported by the
// socket.io server.
func (session *Session) SupportProtocol(protocol string) bool {
	for _, supportedProtocol := range session.SupportedProtocols {
		if protocol == supportedProtocol {
			return true
		}
	}
	return false
}
