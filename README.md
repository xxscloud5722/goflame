# flame
> 使用Golang 语言开发的监测`网站证书` 以及`网站接口` 健康情况，并将数据在`企业微信推送`

## 配置文件
> application.yaml
```yaml
ssl:
  - name: example
    host: www.example.com
    principal: example

probe:
  - name: example
    http:
      - url: 'https://example.com/data/example'
        method: GET

notice:
  - model: WECHAT
    value: ''
```