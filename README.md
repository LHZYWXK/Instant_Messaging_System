# 即时通信系统
## 生成可执行文件
- server
`go build -o server main.go server.go user.go`
- client
`go build -o client client.go`
## 开启服务器和客户端
- 服务器
在终端内输入`./server`开启服务器
- 客户端
每连接一个用户需要开启一个终端
在终端内输入`./client`开启客户端
## 命令行解析
- 在终端内输入`./client -h`可以查看说明
- 在终端内输入`./client -ip xxx.xxx.xxx.xxx -port xxxx`可自定义地址和端口
