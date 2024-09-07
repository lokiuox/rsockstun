package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	socks5 "github.com/armon/go-socks5"

	"github.com/hashicorp/yamux"
)

var session *yamux.Session

var proxytout = time.Millisecond * 1000 //timeout for wait magicbytes
// Catches yamux connecting to us
func clientListener(address string, certificate string) {
	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: starting socks proxy\n")
		os.Exit(1)
	}

	cer, err := tls.LoadX509KeyPair(certificate+".crt", certificate+".key")
	if err != nil {
		log.Println(err)
		fmt.Fprintf(os.Stderr, "Please check the program's usage on how to generate a new SSL certificate.\n")
		os.Exit(1)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", address, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot listen on %v\n", address)
		os.Exit(1)
	}
	log.Printf("Listening for clients on %v\n", address)
	for {
		conn, err := ln.Accept()
		conn.RemoteAddr()
		log.Printf("Got a SSL connection from %v: ", conn.RemoteAddr())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Errors accepting connection!\n")
		}

		reader := bufio.NewReader(conn)

		//read only 64 bytes with timeout=1-3 sec. So we haven't delay with browsers
		conn.SetReadDeadline(time.Now().Add(proxytout))
		statusb := make([]byte, 64)
		_, _ = io.ReadFull(reader, statusb)

		if string(statusb)[:len(agentpassword)] != agentpassword {
			//do HTTP checks
			log.Printf("Received request: %v", string(statusb[:64]))
			status := string(statusb)
			if strings.Contains(status, " HTTP/1.1") {
				httpresonse := "HTTP/1.1 301 Moved Permanently" +
					"\r\nContent-Type: text/html; charset=UTF-8" +
					"\r\nLocation: https://www.microsoft.com/" +
					"\r\nServer: Apache" +
					"\r\nContent-Length: 0" +
					"\r\nConnection: close" +
					"\r\n\r\n"

				conn.Write([]byte(httpresonse))
				conn.Close()
			} else {
				conn.Close()
			}

		} else {
			//magic bytes received.
			//disable socket read timeouts
			log.Println("Client Connected.")
			conn.SetReadDeadline(time.Now().Add(100 * time.Hour))

			//Add connection to yamux
			session, err = yamux.Server(conn, nil)

			if err != nil {
				log.Println(err)
				fmt.Fprintf(os.Stderr, "yamux error detected.\n")
				conn.Close()
				continue
			}

			for {
				stream, err := session.Accept()
				log.Println("Acceping stream")
				if err != nil {
					conn.Close()
					fmt.Fprintf(os.Stderr, "Errors accepting stream!\n")
					continue
				}
				log.Println("Passing off to SOCKS")
				go func() {
					err = server.ServeConn(stream)
					if err != nil {
						log.Println(err)
					}
				}()
			}

		}
	}
}
