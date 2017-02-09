## acproxy
a proxy interacting with hadoop clusters, which supports:

* proxy webhdfs request to active namenode instead of standby namenode (if send requests to standby namenode, standbyexception will return)

```
go run acproxy.go --type=hdfs --config_file=examples/config.yaml --log_dir=/var/log --v=1 --alsologtostderr=true
# configuration can be set in config.yaml, or overrided by environment variables

Flags:
      --alsologtostderr                  log to standard error as well as files
  -c, --config_file string               location of config file (default "config.yaml")
      --log_dir string                   If non-empty, write log files in this directory
  -t, --type string                      proxy provider type chosen in {hdfs} (default "hdfs")
      --logtostderr                      log to standard error instead of files
  -v, --v Level                          log level for V logs
```

### interfaces

#### 1. ip:port/states
get states of different proxy providers, which temporarily has only three states: **initing, running and pending**
```
 curl ip:port/states
 {
    "provider_state": "running",
    "state_explanation": "hdfs proxy is in service"
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
            "host": "localhost",
            "path": "/webhdfs/v1/",
            "status_code": 200,
            "status": "OK",
            "delay": 10453166
        },
		 {
            "method": "put",
            "host": "localhost",
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
proxy requests
```
curl ip:port/webhdfs/v1/<PATH>?op=LISTSTATUS
curl -X PUT ip:port/webhdfs/v1/<PATH>?op=MKDIRS
...
```
