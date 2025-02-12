# GoDesk Client

`Remote Desktop Client` depend on **GO** 、 **Vue** 、 **Wails**

### Core Components

+ [Wails](https://wails.io)
+ [Golang](https://go.dev/)
+ [Vue 3](https://v3.vuejs.org/)

### Dev

+ 项目运行

```shell
wails dev
```

+ proto文件生成

```shell
protoc -I ./proto --go_out=./proto/ --go_opt=paths=source_relative \
 --go-grpc_out=./proto/ --go-grpc_opt=require_unimplemented_servers=false \
 --go-grpc_opt=paths=source_relative ./proto/*.proto
```

## Build

```shell
wails build
```
