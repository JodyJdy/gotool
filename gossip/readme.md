# 安装依赖
```shell
cd common
go install
```

# build
##  打包test
test 启动了多个gossip节点，测试功能

```shell
cd test
go build
```
## 打包 sendlog

sendlog 向gossip集群发送消息
```shell
cd sendlog
go build
```

# 使用
## test
test.go 有三个参数
* -nodes 启动节点数量
* -port  节点占用的起始端口，节点会占用 [port,port+nodes]范围的端口
* -iface  使用的网卡名称 多播使用到 

```shell

test.exe -nodes 3 -port 8008  -iface "WLAN"

```

## sendlog
sendlog.go 有两个参数
* -port  节点占用的起始端口，节点会占用 [port,port+nodes]范围的端口
* -cur  数据要发送到的节点的下标

```shell
sendlog.exe -port 8080  -cur 0
```
