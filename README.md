# nginx-log-exporter
A Nginx log parser exporter for prometheus metrics.

![screen shot 2018-01-08 at 9 36 21 am](https://user-images.githubusercontent.com/1459834/34656613-7083cf3e-f457-11e7-929a-2758abad387b.png)


## Installation

1. go get `https://github.com/songjiayang/nginx-log-exporter`
2. Or use [binary](https://github.com/songjiayang/nginx-log-exporter/releases) release

## Usage

```
nginx-log-exporter -h 

Usage of:
  -config.file string
    	Nginx log exporter configuration file name. (default "config.yml")
  -web.listen-address string
    	Address to listen on for the web interface and API. (default ":9999")
exit status 2
```

## Configuration

```
- name: app
  format: $remote_addr - $remote_user [$time_local] "$method $request $protocol" $request_time-$upstream_response_time $status $body_bytes_sent "$http_referer" "$http_user_agent" "$http_x_forwarded_for"
  source_files:
    - ./test/nginx.log
  static_config:
    foo: foo
  relabel_config: 
    source_labels: 
      - request
      - method
      - status
    replacement:
      request: 
        trim: "?"
        replace:
          - target: /v1.0/example/\d+
            value: /v1.0/example/:id
```

- format: your nginx `log_format` regular expression, notice: you should make a new one for your app.
- name: service name, metric will be `{name}_http_response_count_total`, `{name}_http_response_count_total`, `{name}_http_response_size_bytes`, `{name}_http_upstream_time_seconds`, `{name}_http_response_time_seconds`
- source_files: sevice nginx log, support multiple files.
- static_config: all metrics will add static labelsets.
- relabel_config:
  * source_labels: what's labels should be use.
  * replacement: source labelvalue format rule, it supports regrex, eg `/v1.0/example/123?id=q=xxx` will relace to `/v1.0/example/:id`, it's very powerful. 

## Example

After parse `./test/nginx.log`, the result is

```
# HELP app_http_response_count_total Amount of processed HTTP requests
# TYPE app_http_response_count_total counter
app_http_response_count_total{foo="foo",method="GET",request="/v1.0/example",status="200"} 2
app_http_response_count_total{foo="foo",method="GET",request="/v1.0/example/:id",status="200"} 1
# HELP app_http_response_size_bytes Total amount of transferred bytes
# TYPE app_http_response_size_bytes counter
app_http_response_size_bytes{foo="foo",method="GET",request="/v1.0/example",status="200"} 70
app_http_response_size_bytes{foo="foo",method="GET",request="/v1.0/example/:id",status="200"} 21
# HELP app_http_response_time_seconds Time needed by NGINX to handle requests
# TYPE app_http_response_time_seconds histogram
app_http_response_time_seconds_bucket{foo="foo",method="GET",request="/v1.0/example",status="200",le="0.005"} 2
.....
app_http_response_time_seconds_count{foo="foo",method="GET",request="/v1.0/example",status="200"} 2
app_http_response_time_seconds_bucket{foo="foo",method="GET",request="/v1.0/example/:id",status="200",le="0.005"} 1

app_http_response_time_seconds_sum{foo="foo",method="GET",request="/v1.0/example/:id",status="200"} 0.003
app_http_response_time_seconds_count{foo="foo",method="GET",request="/v1.0/example/:id",status="200"} 1
# HELP app_http_upstream_time_seconds Time needed by upstream servers to handle requests
# TYPE app_http_upstream_time_seconds histogram

```

## Thanks

- Inspired by [prometheus-nginxlog-exporter](https://github.com/martin-helmich/prometheus-nginxlog-exporter)
