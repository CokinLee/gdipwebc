# gdipwebc

Yet another [GnuDIP][gnudip] http client implementation in Go

[gnudip]: http://gnudip2.sourceforge.net/

## What is GnuDIP

GnuDIP is a software that implements a Dynamic DNS service.
Development/maintenance of GnuDIP is stopped for a long time.

## What is gdipwebc

gdipwebc is a GnuDIP client implementation in Go.  It only supports
[the HTTP based protocol][protocol].  It is [dockerized][dockerimg]
and easily deployable on kubernetes clusters.

[protocol]: http://gnudip2.sourceforge.net/gnudip-www/latest/gnudip/html/protocol.html
[dockerimg]: https://gitlab.com/ktateish/gdipwebc/container_registry

## Usage

### CLI

* Install
```
$ go get github.com/ktateish/gdipwebc
```
* Register the address that the server see me at, and pass it back to me
```
$ gdipwebc --url http://gnudip.svc.example.com/ --user foo --password my_secret_pw --domain-name foo.example.org
192.168.1.1
```
* Register a specific address
```
$ gdipwebc --url http://gnudip.svc.example.com/ --user foo --password my_secret_pw --domain-name foo.example.org --address 192.168.1.1
192.168.1.1
```

### Docker

TBU

### Kubernetes

TBU

## Acknowledgement

* [GnuDIP][gnudip]
* [GnuDIP Protocol][protocol]

## Author

Katsuyuki Tateishi <kt@wheel.jp>
