## active-proxy
a proxy aims to interacting with hadoop clusters, seems like API Gateway, which supports:

* proxy webhdfs request to active namenode instead of standby namenode (if send requests to standby namenode, standbyexception will return)

```
go run server.go -config="examples/config.yaml"
# configuration can be set in config.yaml, or overrided by environment variables
```

### interfaces

#### 1. ip:port/states
get states of different proxy providers
```
 curl ip:port/states
 {
    "hdfs_proxy_provider": {
        "provider_state": "start",
        "state_explanation": "hdfs proxy is in service"
    }
 }
```

#### 2. ip:port/statistics
get some statistics and recent request records (including delay, statuscode and so on)
```
 curl ip:port/statistics
 {
    "recentRequests": [
        {
            "method": "GET",
            "host": "localhost:8080",
            "path": "/webhdfs/v1/",
            "status_code": 200,
            "status": "OK",
            "delay": 10453166
        },
	{
            "method": "put",
            "host": "localhost:8080",
            "path": "/webhdfs/v1/tmp/test",
            "status_code": 403,
            "status": "Forbidden",
            "delay": 2045863753
        }
    ],
    "totalRequests": 2
 }
```

#### 3. ip:port/*
proxy requests according to some specific rules:

* hdfs proxy provider proxy for request uri starting with "/webhdfs/v1"
```
curl ip:port/webhdfs/v1/<PATH>?op=LISTSTATUS
curl -X PUT ip:port/webhdfs/v1/<PATH>?op=MKDIRS
...
```
