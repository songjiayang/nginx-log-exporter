# nginx-log-exporter
A Exporter to parse Nginx/gin request log to Prometheus metrics.

![screen shot 2018-01-08 at 9 36 21 am](https://user-images.githubusercontent.com/1459834/34656613-7083cf3e-f457-11e7-929a-2758abad387b.png)


## Features

- Various log formats support.
- Request path can be rewrited.
- Status code can be mergeable.
- Prometheus Exemplar support.


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
  -placeholder.replace bool
         Whether enable placeholder replacement when rewriting the request path  (default false)
exit status 2
```

## Configuration

```
- name: nginx
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
      status:
        replace:
          - target: 4.+
            value: 4xx
          - target: 5.+
            value: 5xx
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

- name: gateway_uri
  format: $remote_addr - $remote_user [$time_local] "$method $uri $protocol" $request_time-$upstream_response_time $status $body_bytes_sent "$http_referer" "$http_user_agent" "$http_x_forwarded_for" $request_id

  source_files:
    - ./logs/nginx_gateway.access.log
  relabel_config:
    source_labels:
      - uri
      - method
      - status
    replacement:
      uri:
        trim: "?"
        replace:
          # replace url "/api/path1/path2/123/xxx" to  "/api/path1/path2/:id/xxx"
          # replace url "/api/path7/path8/789/xxx" to  "/api/path7/path8/:id/xxx"
          - target: (^\/api\/[^\/]+\/[^\/]+\/)\d+(.*)
            value: ${1}:id${2}
          # replace url "/api/path1/path2/path3/123/xxx" to  "/api/path1/path2/path3/:id/xxx"
          # replace url "/api/path7/path8/path9/789/xxx" to  "/api/path7/path8/path9/:id/xxx"
          - target: (^\/api\/[^\/]+\/[^\/]+\/[^\/]+\/)\d+(.*)
            value: ${1}:id${2}
          - target: (^\/api\/[^\/]+\/[^\/]+\/[^\/]+\/)\d+(.*)
            value: ${1}:id${2}
          # The remaining uris starting with /api/ remain as is
          - target: ^\/api\/.*
            value: ${0}
          # The uri that does not match the above rules and is not my concern, please replace it with 'notATargetUri'
          - target: (.*)
            value: "notATargetUri"
      status:
        replace:
          - target: 404
            value: 404
          - target: 4.+
            value: 4xx
          - target: 5.+
            value: 5xx
  histogram_buckets: [0.1, 0.3, 0.5, 1, 2]

- name: gateway_service
  format: $remote_addr - $remote_user [$time_local] "$method $service $protocol" $request_time-$upstream_response_time $status $body_bytes_sent "$http_referer" "$http_user_agent" "$http_x_forwarded_for" $request_id

  source_files:
    - ./logs/nginx_gateway.access.log
  relabel_config:
    source_labels:
      - method
      - status
      - service
    replacement:
      service:
        trim: "?"
        replace:
          # extract the second part from the request_uri as the service name.
          # The service of "/api/user/path1/123" is "user", and the service of "/api/order/path1/123" is "order"
          # Then we can generate metrics for the service dimension
          - target: \/api\/([\d\w-]+)
            value: ${1}
          - target: (.*)
          - value: ""
      status:
        replace:
          - target: 404
            value: 404
          - target: 4.+
            value: 4xx
          - target: 5.+
            value: 5xx
  histogram_buckets: [0.1, 0.3, 0.5, 1, 2]
```

- format: your nginx `log_format` regular expression, notice: you should make a new one for your app, variable your log with format configuration, you almost have some variables like `body_bytes_sent`, `upstream_response_time`, `request_time`.
- name: service name, metric will be `{name}_http_response_count_total`, `{name}_http_response_count_total`, `{name}_http_response_size_bytes`, `{name}_http_upstream_time_seconds`, `{name}_http_response_time_seconds`
- source_files: service nginx log, support multiple files.
- external_labels: all metrics will add this labelsets.
- relabel_config:
  * source_labels: what's labels should be use.
  * replacement: source labelvalue format rule, it supports regrex, eg `/v1.0/example/123?id=q=xxx` will relace to `/v1.0/example/:id`, you also can use it to merge 400, 429 to 4xx. 
- histogram_buckets: configure histogram metrics buckets.
- exemplar_config: configure exemplars, it used for histogram metrics.
 
## Example

`./test/nginx.log` output

```
# HELP nginx_http_response_count_total Amount of processed HTTP requests
# TYPE nginx_http_response_count_total counter
nginx_http_response_count_total{method="GET",region="zone1",request="/v1.0/example",status="200"} 2
nginx_http_response_count_total{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 1
nginx_http_response_count_total{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx"} 1
nginx_http_response_count_total{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx"} 1
# HELP nginx_http_response_size_bytes Total amount of transferred bytes
# TYPE nginx_http_response_size_bytes counter
nginx_http_response_size_bytes{method="GET",region="zone1",request="/v1.0/example",status="200"} 70
nginx_http_response_size_bytes{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 21
nginx_http_response_size_bytes{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx"} 21
nginx_http_response_size_bytes{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx"} 21
# HELP nginx_http_response_time_seconds Time needed by NGINX to handle requests
# TYPE nginx_http_response_time_seconds histogram
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.1"} 2
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.3"} 2
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.5"} 2
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="1"} 2
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="2"} 2
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="+Inf"} 2
nginx_http_response_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example",status="200"} 0.005
nginx_http_response_time_seconds_count{method="GET",region="zone1",request="/v1.0/example",status="200"} 2
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.1"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.3"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.5"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="1"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="2"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="+Inf"} 1
nginx_http_response_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 0.003
nginx_http_response_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="0.1"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="0.3"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="0.5"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="1"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="2"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="+Inf"} 1
nginx_http_response_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx"} 0.003
nginx_http_response_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="0.1"} 0
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="0.3"} 0
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="0.5"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="1"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="2"} 1
nginx_http_response_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="+Inf"} 1
nginx_http_response_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx"} 0.5
nginx_http_response_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx"} 1
# HELP nginx_http_upstream_time_seconds Time needed by upstream servers to handle requests
# TYPE nginx_http_upstream_time_seconds histogram
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.1"} 2
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.3"} 2
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="0.5"} 2
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="1"} 2
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="2"} 2
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example",status="200",le="+Inf"} 2
nginx_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example",status="200"} 0.005
nginx_http_upstream_time_seconds_count{method="GET",region="zone1",request="/v1.0/example",status="200"} 2
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.1"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.3"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="0.5"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="1"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="2"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="200",le="+Inf"} 1
nginx_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 0.003
nginx_http_upstream_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="200"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="0.1"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="0.3"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="0.5"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="1"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="2"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx",le="+Inf"} 1
nginx_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx"} 0.003
nginx_http_upstream_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="4xx"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="0.1"} 0
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="0.3"} 0
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="0.5"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="1"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="2"} 1
nginx_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx",le="+Inf"} 1
nginx_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx"} 0.4
nginx_http_upstream_time_seconds_count{method="GET",region="zone1",request="/v1.0/example/:id",status="5xx"} 1
....
```

`./test/gin.log` output

```
gin_http_response_count_total{method="GET",region="zone1",request="/ping",status="200"} 1
gin_http_response_count_total{method="GET",region="zone1",request="/users",status="200"} 1
# HELP gin_http_upstream_time_seconds Time needed by upstream servers to handle requests
# TYPE gin_http_upstream_time_seconds histogram
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="0.1"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="0.3"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="0.5"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="1"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="2"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/ping",status="200",le="+Inf"} 1
gin_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/ping",status="200"} 0.000122767
gin_http_upstream_time_seconds_count{method="GET",region="zone1",request="/ping",status="200"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="0.1"} 0
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="0.3"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="0.5"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="1"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="2"} 1
gin_http_upstream_time_seconds_bucket{method="GET",region="zone1",request="/users",status="200",le="+Inf"} 1
gin_http_upstream_time_seconds_sum{method="GET",region="zone1",request="/users",status="200"} 0.2
gin_http_upstream_time_seconds_count{method="GET",region="zone1",request="/users",status="200"} 1
```

`gateway_uri namespace ` output:
```
# TYPE gateway_uri_http_response_count_total counter
gateway_uri_http_response_count_total{method="GET",status="200",uri="/api/order/path1/:id"} 2
gateway_uri_http_response_count_total{method="GET",status="200",uri="/api/user/path1/:id"} 2
gateway_uri_http_response_count_total{method="POST",status="200",uri="/api/order/path1/path2/:id"} 1
gateway_uri_http_response_count_total{method="POST",status="200",uri="/api/user/path1/path2/:id"} 2
# HELP gateway_uri_http_response_size_bytes Total amount of transferred bytes
# TYPE gateway_uri_http_response_size_bytes counter
gateway_uri_http_response_size_bytes{method="GET",status="200",uri="/api/order/path1/:id"} 70
gateway_uri_http_response_size_bytes{method="GET",status="200",uri="/api/user/path1/:id"} 70
gateway_uri_http_response_size_bytes{method="POST",status="200",uri="/api/order/path1/path2/:id"} 21
gateway_uri_http_response_size_bytes{method="POST",status="200",uri="/api/user/path1/path2/:id"} 42
```

`gateway_service namespace ` output:
```
# TYPE gateway_service_http_response_count_total counter
gateway_service_http_response_count_total{method="GET",service="order",status="200"} 2
gateway_service_http_response_count_total{method="GET",service="user",status="200"} 2
gateway_service_http_response_count_total{method="POST",service="order",status="200"} 1
gateway_service_http_response_count_total{method="POST",service="user",status="200"} 2
# HELP gateway_service_http_response_size_bytes Total amount of transferred bytes
# TYPE gateway_service_http_response_size_bytes counter
gateway_service_http_response_size_bytes{method="GET",service="order",status="200"} 70
gateway_service_http_response_size_bytes{method="GET",service="user",status="200"} 70
gateway_service_http_response_size_bytes{method="POST",service="order",status="200"} 21
gateway_service_http_response_size_bytes{method="POST",service="user",status="200"} 42
# HELP gateway_service_http_response_time_seconds Time needed by NGINX to handle requests
```

## Thanks

- Inspired by [prometheus-nginxlog-exporter](https://github.com/martin-helmich/prometheus-nginxlog-exporter)
