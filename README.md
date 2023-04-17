

## install
```shell
    # install protoc
    # then
    go install github.com/ml444/gctl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
```

## config
```shell
export GCTL_TEMPLATES_ROOT_DIR="/your_path/github.com/ml444/gctl/templates"
export GCTL_MODULE_PREFIX="github.com/xxx"
export GCTL_TARGET_ROOT_PATH="/your/target/root/path"
export GCTL_SVC_GROUP_INIT_ERRCODE_MAP={"base": 20000, "spp": 100000}
```
**Options config**
```shell
export GCTL_SVC_GROUP_INIT_PORT_MAP={"base": 9000, "spp": 10000}
export GCTL_ONCE_FILES=".gitignore go.mod .editorconfig README.md Makefile Dockerfile"
```


## 错误码和服务端口的分配
### 错误码
在创建服务的proto文件时，可以启用自动分配错误码范围的功能，这样可以在生成proto时，
初步展示第一个错误码，并且标注了该服务的错误码范围。

### 端口分配
有时我们可能需要为每个服务设置不同的端口，以满足在本地或者同一个机器上同时运行；
并且每个服务可能需要多个端口来监听不同的业务或功能。


## reading from config file
在用户根目录下创建配置文件：`~/.gctl_config.yaml`
```yaml
TEMPLATES_ROOT_DIR: "/your_path/github.com/ml444/gctl/templates"
MODULE_PREFIX: "github.com/xxx"
TARGET_ROOT_PATH: "/your/target/root/path"

SVC_GROUP_INIT_ERRCODE_MAP: {"base": 20000, "spp": 100000}
SVC_GROUP_INIT_PORT_MAP: {"base": 9000, "spp": 10000}
ONCE_FILES: ".gitignore go.mod .editorconfig README.md Makefile Dockerfile"
```
