# nginx-log-exporter
A Nginx log parser exporter for prometheus metrics.

![screen shot 2018-01-08 at 9 36 21 am](https://user-images.githubusercontent.com/1459834/34656613-7083cf3e-f457-11e7-929a-2758abad387b.png)


## Installation

1. go get `github.com/songjiayang/nginx-log-exporter`
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
  format: $remote_addr - $remote_user [$time_local] "$method $request $protocol" $request_time-$upstream_response_time $status $body_bytes_sent "$http_referer" "$http_user_agent" "$http_x_forwarded_for" $request_id
  source_files:
    - ./test/nginx.log
  external_labels:
    region: zone1
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
  histogram_buckets: [0.1, 0.3, 0.5, 1, 2]
  exemplar_config:
    match:
      request_time: ">= 0.3"
    labels:
      - request_id
      - remote_addr
- name: gin
  format: $clientip - [$time_local] "$method $request $protocol $status $upstream_response_time "$http_user_agent" $err"
  source_files:
    - ./test/gin.log
  external_labels:
    region: zone1
  relabel_config:
    source_labels:
      - request
      - method
      - status
    replacement:
      request:
        trim: "?"
  histogram_buckets: [0.1, 0.3, 0.5, 1, 2]
```

- format: your nginx `log_format` regular expression, notice: you should make a new one for your app, variable your log with format configuration, you almost have some variables like `body_bytes_sent`, `upstream_response_time`, `request_time`.
- name: service name, metric will be `{name}_http_response_count_total`, `{name}_http_response_count_total`, `{name}_http_response_size_bytes`, `{name}_http_upstream_time_seconds`, `{name}_http_response_time_seconds`
- source_files: service nginx log, support multiple files.
- external_labels: all metrics will add this labelsets.
- relabel_config:
  * source_labels: what's labels should be use.
  * replacement: source labelvalue format rule, it supports regrex, eg `/v1.0/example/123?id=q=xxx` will relace to `/v1.0/example/:id`, it's very powerful. 
- histogram_buckets: configure histogram metrics buckets.
- exemplar_config: configure exemplars, it used for histogram metrics.
 
## Example

`./test/nginx.log` result is

```
# HELP app_http_response_count_total Amount of processed HTTP requests
# TYPE app_http_response_count_total counter
app_http_response_count_total{method="GET",region="zone1",request="/v1.0/example",status="200"} 2
app_http_response_count_total{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 1
app_http_response_count_total{method="GET",region="zone1",request="/v1.0/example/:id",status="400"} 1
app_http_response_count_total{method="GET",region="zone1",request="/v1.0/example/:id",status="500"} 1
# HELP app_http_response_size_bytes Total amount of transferred bytes
# TYPE app_http_response_size_bytes counter
app_http_response_size_bytes{method="GET",region="zone1",request="/v1.0/example",status="200"} 70
app_http_response_size_bytes{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 21
app_http_response_size_bytes{method="GET",region="zone1",request="/v1.0/example/:id",status="400"} 21
app_http_response_size_bytes{method="GET",region="zone1",request="/v1.0/example/:id",status="500"} 21
# HELP app_http_response_time_seconds Time needed by NGINX to handle requests
# TYPE app_http_response_time_seconds histogram
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.1"} 2
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.3"} 2
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.5"} 2
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="1"} 2
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="2"} 2
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="+Inf"} 2
app_http_response_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example",status="200"} 0.005
app_http_response_time_seconds_count{method="GET",region="zone1",request="/v1.0/example",status="200"} 2
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.1"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.3"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.5"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="1"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="2"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="+Inf"} 1
app_http_response_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 0.003
app_http_response_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="0.1"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="0.3"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="0.5"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="1"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="2"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="+Inf"} 1
app_http_response_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="400"} 0.003
app_http_response_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="400"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="0.1"} 0
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="0.3"} 0
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="0.5"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="1"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="2"} 1
app_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="+Inf"} 1
app_http_response_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="500"} 0.5
app_http_response_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="500"} 1
# HELP app_http_upstream_time_seconds Time needed by upstream servers to handle requests
# TYPE app_http_upstream_time_seconds histogram
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.1"} 2
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.3"} 2
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.5"} 2
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="1"} 2
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="2"} 2
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="+Inf"} 2
app_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example",status="200"} 0.005
app_http_upstream_time_seconds_count{method="GET",region="zone1",request="/v1.0/example",status="200"} 2
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.1"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.3"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.5"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="1"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="2"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="+Inf"} 1
app_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 0.003
app_http_upstream_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="0.1"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="0.3"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="0.5"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="1"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="2"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="400",le="+Inf"} 1
app_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="400"} 0.003
app_http_upstream_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="400"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="0.1"} 0
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="0.3"} 0
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="0.5"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="1"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="2"} 1
app_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="500",le="+Inf"} 1
app_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="500"} 0.4
app_http_upstream_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="500"} 1
# HELP gin_http_response_count_total Amount of processed HTTP requests
# TYPE gin_http_response_count_total counter
gin_http_response_count_total{method="GET",region="zone1",request="/ping",status="200"} 2
gin_http_response_count_total{method="GET",region="zone1",request="/users",status="200"} 2
# HELP gin_http_upstream_time_seconds Time needed by upstream servers to handle requests
# TYPE gin_http_upstream_time_seconds histogram
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="0.1"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="0.3"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="0.5"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="1"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="2"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="+Inf"} 2
gin_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/ping",status="200"} 0.000245534
gin_http_upstream_time_seconds_count{method="GET",region="zone1",request="/ping",status="200"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="0.1"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="0.3"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="0.5"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="1"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="2"} 2
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="+Inf"} 2
gin_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/users",status="200"} 0.200122767
gin_http_upstream_time_seconds_count{method="GET",region="zone1",request="/users",status="200"} 2
```

## Thanks

- Inspired by [prometheus-nginxlog-exporter](https://github.com/martin-helmich/prometheus-nginxlog-exporter)
