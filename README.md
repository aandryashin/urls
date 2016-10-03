# Urls
[![Build Status](https://travis-ci.org/aandryashin/urls.svg?branch=master)](https://travis-ci.org/aandryashin/urls)
[![Coverage](https://codecov.io/github/aandryashin/urls/coverage.svg)](https://codecov.io/gh/aandryashin/urls)

Simpliest Distributed Url Shortener

## Building
Use [godep](https://github.com/tools/godep) for dependencies management so ensure it's installed before proceeding with next steps. To build the code:

1. Checkout this source tree: ```$ git clone https://github.com/aandryashin/urls.git```
2. Download dependencies: ```$ godep restore```
3. Build as usually: ```$ go build```
4. Run compiled binary: ```./urls --help```

## Running
As distributed storage Urls uses etcd daemon, so run it first in separate terminal, then run urls:

```
$ etcd
$ ./urls
2016/10/02 21:26:51 Serving [::]:8080 with pid 8132
```

Check it with:

```
$ curl http://localhost:8080 -d'{"url" : "http://www.google.com"}'
{"url":"4"}
$ curl http://localhost:8080/4
<a href="http://www.google.com">Moved Permanently</a>.
```

Enable https support:

```
$ ./urls -http :8080 -https :8443
2016/10/03 10:07:35 Serving [::]:8080, [::]:8443 with pid 11660 
```

```
$ curl -L -k http://localhost:8080 -d'{"url" : "http://www.google.com"}'
{"url":"8"}
$ curl -k http://localhost:8080/8
<a href="https://[::]:8443/8">Temporary Redirect</a>.
```

To run with several etcd nodes, you have to provide them with endpoints option.
