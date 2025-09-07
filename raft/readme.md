# build

## 1. 构建test
test.go 可以启动一个raft节点,用于测试raft功能
```shell
cd test

go build 
```

## 2. 构建 sendlog
sendlog.go 用于周期性的向leader节点发送消息,测试消息同步
```shell
cd sendlog
go build
```

# test使用
test.go 有三个命令行参数:
* -nodes 指定节点数量
* -port  指定节点占用的起始端口，每个节点占用一个端口
* -cur 当前节点下标

以windows平台为例，如果要启动三个节点，可以启动三个终端并执行以下命令:

```shell
test.exe -nodes 3 -port 8000 -cur 0 
```
```shell
test.exe -nodes 3 -port 8000 -cur 1
```

```shell
test.exe -nodes 3 -port 8000 -cur 2
```
那么会启动三个节点，分别占用端口: 8000,8001,8002

# sendlog 使用
启动节点后，可以使用sendlog向leader节点发送消息,sendlog 有一个参数:
* -leader  指定leader节点的地址
使用方式如下：

```shell
 sendlog.exe -leader ":8000"
```