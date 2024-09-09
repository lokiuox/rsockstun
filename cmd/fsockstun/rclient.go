package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/url"

	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/launchdarkly/go-ntlm-proxy-auth"
)

var encBase64 = base64.StdEncoding.EncodeToString
var decBase64 = base64.StdEncoding.DecodeString
var username string = ""
var domain string = ""
var password string = ""
var connectproxystring string
var useragent string
var proxytimeout = time.Millisecond * 1000 //timeout for proxyserver response

func getProxyConnection(proxyaddr string, connectaddr string) net.Conn {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	proxyurl, err := url.Parse(proxyaddr)
	if err != nil {
		log.Println("url parse error:", err)
		return nil
	}

	ntlmDialContext := ntlm.NewNTLMProxyDialContext(dialer, *proxyurl, username, password, domain, nil)
	if ntlmDialContext == nil {
		log.Println("ntlmDialErr")
		return nil
	}
	ctx := context.Background()
	conn, err := ntlmDialContext(ctx, "tcp", connectaddr)
	if err != nil {
		log.Println("ntlm dialContext connection error:", err)
		return nil
	}

	return conn
}

func connectToServer(address string, proxy string, socks string, serverName string) error {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	if serverName != "" {
		conf.ServerName = serverName
	}

	var err error
	var conn net.Conn
	var connp net.Conn
	var session *yamux.Session
	tryToReconnect := true
 
	for tryToReconnect {
	if proxy == "" {
		log.Println("Connecting to far end")
		conn, err = tls.Dial("tcp", address, conf)
		if err != nil {
			return err
		}
	} else {
		log.Println("Connecting to proxy ...")
		connp = getProxyConnection(proxy, address)
		if connp != nil {
			log.Println("Proxy connection successful. Connecting to far end...")
			conntls := tls.Client(connp, conf)
			err := conntls.Handshake()
			if err != nil {
				log.Printf("Error connecting: %v", err)
				return err
			}
			conn = net.Conn(conntls)
		} else {
			log.Println("Proxy connection NOT successful. Exiting")
			return nil
		}
	}

	log.Println("Starting client")
	conn.Write([]byte(agentpassword))
	time.Sleep(time.Second * 1)
	session, err = yamux.Client(conn, nil)
	if err != nil {
		return err
	}

	tryToReconnect, err = socksListenerClient(socks, session)
}
	return err
}

// Catches clients and connects to yamux
func socksListenerClient(address string, session *yamux.Session) (bool, error) {
	ln, err := net.Listen("tcp", address)
	tryToReconnect := false
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot start SOCKS on %v\n", address)
		return tryToReconnect, err
	}
	log.Printf("Listening for SOCKS connections on %v\n", address)
	shouldReturn := false
	for !shouldReturn {
		conn, err := ln.Accept()
		go func() {
			if err != nil {
				log.Println("Accepting SOCKS stream error")
				return
			}

			log.Println("Got a connection, opening a stream...")

			stream, err := session.Open()
			if err != nil {
				log.Println("Opening SOCKS session error")
				shouldReturn = true
				tryToReconnect = true
				return
			}

			// connect both of conn and stream
			go func() {
				log.Println("Starting to copy conn to stream")
				io.Copy(conn, stream)
				conn.Close()
			}()
			go func() {
				log.Println("Starting to copy stream to conn")
				io.Copy(stream, conn)
				stream.Close()
				log.Println("Done copying stream to conn")
			}()
		}()
	}
	return tryToReconnect, nil
}
