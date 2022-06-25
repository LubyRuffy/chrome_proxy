# chrome_docker

用chrome的docker环境做最简单的截图服务器。

## 编译

```shell
docker build --tag lubyruffy/chrome_proxy:latest .
```

## 测试运行

运行
```shell
docker run --rm -it -p5558:5558 lubyruffy/chrome_proxy:latest
```

截图
```shell
curl -d '{"url":"http://www.baidu.com", "sleep":1, "timeout":10}' http://127.0.0.1:5558/screenshot
```
```json
{
  "code": 200,
  "url": "http://www.baidu.com",
  "data": "iVB...base64..."
}
```

渲染dom
```shell
curl -d '{"url":"http://www.baidu.com", "sleep":1, "timeout":10}' http://127.0.0.1:5558/renderDom
```
```json
{
  "code": 200,
  "url": "http://www.baidu.com",
  "data": "<html>...</html>"
}
```