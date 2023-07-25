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

保存镜像
```shell
docker save -o chrome_proxy.tar lubyruffy/chrome_proxy:latest
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

附带url的截图
```shell
curl -d '{"url":"http://www.baidu.com", "sleep":1, "timeout":10, "add_url": true}' http://127.0.0.1:5558/screenshot
```
```json
{
    "code": 200,
    "url": "http://www.baidu.com",
    "data": "/9j/4...base64..."
}
```
截图预览：
![screenshot_with_url.png](screenshot_with_url.png)

渲染dom
```shell
curl -d '{"url":"http://www.baidu.com", "sleep":1, "timeout":10}' http://127.0.0.1:5558/renderDom
```
```json
{
  "code": 200,
  "url": "https://www.baidu.com",
  "data": "<html>...</html>",
  "title": "百度一下，你就知道",
  "location": "https://www.baidu.com/"
}
```

使用自定义代理 & UA
```shell
curl -d '{"url":"http://www.baidu.com", "sleep":1, "timeout":10, "proxy":"socks5://127.0.0.1:7890", "user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.124 Safari/537.36 Edg/102.0.1245.44"}' http://127.0.0.1:5558/screenshot
```
```json
{
  "code": 200,
  "message": "",
  "url": "http://www.baidu.com",
  "data": "iVBOR..."
}
```

