**基于 golang 实现的兼容不蒜子接口的服务**

首先感谢 [bruce sha](http://ibruce.info/about/) 同学开发出 [不蒜子](http://ibruce.info)
免费给大家使用，现在的很多主题都支持不蒜子统计，鉴于不蒜子官方的接口有时候压力太大不稳定，于是就有了私有部署的想法, 于是就有了本项目

Github项目地址 [wangdan7245/busuanzi-golang](https://github.com/wangdan7245/busuanzi-golang)

# 功能：

- 兼容不蒜子接口 v2.4
- 可以私有部署，支持docker
- 使用redis作为数据库，UA统计使用HyperLogLog[1]去重
- 自动生成 js 文件

> [1] 基于 HyperLogLog 去重会损失部分精度，但是个人网站访问量不会太大，精度足够满足日常统计需求，加之其存储空间小去重效率高的优点，个人感觉用它是极好的
# Docker 部署

使用 docker-compose

创建 docker-compose.yml 文件

```yaml
version: '3'
services:
  redis:
    image: redis:6.0-alpine
    container_name: bsz-redis
    command: redis-server --save 60 1 --loglevel warning
    restart: always
    volumes:
      - ./data:/data
  server:
    image: vincent7681/busuanzi-golang
    container_name: bsz-server
    restart: always
    ports:
      - 18080:18080
    environment:
      REDIS_HOST: redis:6379
      DOMAIN: __YOU_DOMAIN__
```

修改 `__YOU_DOMAIN__` 为你的服务器域名即可，会自动生成带有域名的js文件
执行部署和启动命令

```shell
 docker-compose up -d
```

直接访问 `你的IP:18080/busuanzi.pure.mini.js`，或者通过Nginx反代18080端口后，使用域名访问 `你的域名/busuanzi.pure.mini.js`

即可获得用于前端的js文件，可将其替换原来的不蒜子js文件，或者将不蒜子 js 文件的引用链接替换为上面的链接即可正常使用

# Nginx 反代配置

添加如下配置即可

```nginx
    location  /
    {
        proxy_pass http://127.0.0.1:18080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header REMOTE-HOST $remote_addr;
        add_header X-Cache $upstream_cache_status;
    }
```
