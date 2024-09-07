fsockstun
======
[![Go](https://github.com/lokiuox/rsockstun/actions/workflows/go.yml/badge.svg)](https://github.com/lokiuox/rsockstun/actions/workflows/go.yml)    
Create a reverse SOCKS5 proxy via an SSL Tunnel, with support for HTTP proxies.    
Forked from https://github.com/llkat/rsockstun    
Based on https://github.com/brimstone/rsocks    

Usage:
------
```
fsockstun - reverse socks5 server/client

SERVER MODE:
./fsockstun server [-listen <listenAddr>] [-socks <socksAddr>] [options]
Options:
  -cert string
    	Server certificate file prefix (default "server")
  -listen string
    	Listen address for client connections (default "0.0.0.0:8080")
  -socks string
    	Listen address for the SOCKS5 proxy (default "127.0.0.1:1080")

CLIENT MODE:
./fsockstun client -connect <connectAddr> [-proxy <proxyURI>] [options]
Options:
  -connect string
    	address:port of the server to connect to
  -proxy string
    	URI of the proxy to use to connect to the server [optional]
  -proxyauth string
    	Proxy authentication in the format [Domain/]Username:Password [optional]
  -proxytimeout int
    	Proxy response timeout in seconds [optional] (default 1)
  -recn int
    	Reconnection limit, 0 for infinite [optional] (default 3)
  -rect int
    	Reconnection delay [optional] (default 30)
  -useragent string
    	User-Agent [optional] (default "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko")

General Options:
  -pass string
    	Shared server/client password [optional]
  -version
    	Version information

Note: you can generate a new server certificate with the following command:
openssl req -new -x509 -keyout server.key -out server.crt -days 365 -nodes

 ```

## Compilation:
```
Linux
- git clone https://github.com/lokiuox/fsockstun
- cd fsockstun
- go build

Windows client:
- same as linux
- optional: to build as Windows GUI: go build -ldflags -H=windowsgui
- optional: to compress exe - use any exe packer, ex: UPX

launch server:
./fsockstun -listen :8443 -socks 127.0.0.1:1080 -cert cert -agentpassword Password1234

launch client:
./fsockstun -connect clientIP:8443 -agentpassword Password1234 -proxy proxy.domain.local:3128 -proxyauth Domain\userpame:userpass -useragent "Mozilla 5.0/IE Windows 10"

Client connects to server and send agentpassword to authorize on server. If server does not receive agentpassword or reveive wrong pass from client (for example if spider or client browser connects to server ) then it send HTTP 301 redirect code to www.microsoft.com
```
