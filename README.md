# OhttpsWebhook

![GitHub all releases](https://img.shields.io/github/downloads/sgpublic/OhttpsWebhook/total) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/sgpublic/OhttpsWebhook) ![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/sgpublic/OhttpsWebhook?include_prereleases)

这是一个基于 Webhook 的、适用于 [ohttps.com](https://ohttps.com/) 的开源自动化部署工具。

## 食用方法

前往 [OHTTPS - 部署节点创建](https://ohttps.com/guide/createcloudserver) 创建 Webhook 部署节点，并获取回调令牌。

从 [Release](https://github.com/sgpublic/OhttpsWebhook/releases) 下载适合您服务器架构的版本，创建一个配置文件，模板如下：

```yaml
hook:
  path: "/ohttps" # 监听路径，默认为 /ohttps
  listen: "0.0.0.0:8081" # 监听 IP 和端口，默认为 0.0.0.0:8081
config:
  key: "9...c" # ohttps.com 生成的回调令牌
  logging:
    path: "/var/log/ohttps/" # 日志输出目录，默认为 ./log
    aging: 259200 # 日志保留期限，单位：秒，默认为 259200
  nginx-reload-command: "nginx -s reload" # 设置 nginx 重新加载命令，默认为 nginx -s reload
targets:
  - domain: "*.example1.com" # 证书域名
    cert-key: "/etc/nginx/cert/example1.com.key" # 私钥证书保存路径
    fullchain-certs: "/etc/nginx/cert/example1.com.pem" # 证书文件（包含证书和中间证书）保存路径
  - domain: "*.example2.com" # 可添加多个配置
    cert-key: "..."
    fullchain-certs: "..."
```

使用命令行启动，并制定配置文件（若不指定配置文件则尝试读取 `./config.yaml`）：

```shell
ohttps -c /etc/ohttps.d/config.yaml
```

或者（仅在 `Ubuntu` 完成测试）将 [ohttps.service](https://github.com/sgpublic/OhttpsWebhook/blob/master/bin/service/ohttps.service) 文件保存到 `/usr/lib/systemd/system` 目录下，使用 `systemctl` 启动服务（需将配置文件存到 `/etc/ohttps.d/config.yaml`）。

添加 `-s` 参数可以服务模式运行。
