# GCTL
[English](README.md)

一个微服务的代码生成和代码检查的工具。

## 安装
> `gctl`需要用到`protoc`，所以请在开始使用前先安装[protoc](https://github.com/protocolbuffers/protobuf/releases)
```shell
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/ml444/gctl@latest
```

## 配置
1. **通过环境变量来配置**
```shell
# ---必须设置项---
#export GCTL_TEMPLATES_BASE_DIR="/your_path/github.com/ml444/gctl-templates/centralized_templates"
export GCTL_TEMPLATES_BASE_DIR="/your_path/github.com/ml444/gctl-templates/separation_templates"
export GCTL_PROTO_PATHS="/your_path/github.com/ml444/gctl-templates/protofiles"
export GCTL_MODULE_PREFIX="my.gitlab.com"
```
```shell
# ---可选的配置---
export GCTL_TARGET_BASE_DIR="/your/target/root/path"
export GCTL_PROTO_CENTRAL_REPO_PATH="/your_path/my.gitlab.com/my_group/proto"
export GCTL_ONCE_FILES=".gitignore go.mod .editorconfig README.md Makefile Dockerfile"
export GCTL_DEFAULT_SVC_GROUP="my_group"
export GCTL_DB_DSN="mysql://user:password@tcp(localhost:3306)/my_config"
export GCTL_ENABLE_ASSIGN_PORT=false
export GCTL_ENABLE_ASSIGN_ERRCODE=true
export GCTL_SVC_PORT_INTERVAL=5
export GCTL_SVC_ERRCODE_INTERVAL=1000
export GCTL_SVC_GROUP_INIT_PORT_MAP={"base": 9000, "biz": 10000}
export GCTL_SVC_GROUP_INIT_ERRCODE_MAP={"base": 10000, "biz": 100000}
```

2. **通过yaml文件来配置**

在用户根目录下创建配置文件：`~/.gctl_config.yaml`
下面的示例是包含所有配置项，除了必要设置之外，其他的可以根据需要删减。
```yaml
#TemplatesBaseDir: "/your_path/github.com/ml444/gctl-templates/centralized_templates"
TemplatesBaseDir: "/your_path/github.com/ml444/gctl-templates/separation_templates"
TargetBaseDir: "/your/target/root/path"
ModulePrefix: "my.gitlab.com"
DefaultSvcGroup: "my_group"
ProtoPaths: "/your_path/github.com/ml444/gctl-templates/protofiles"
ProtoCentralRepoPath: "/your_path/my.gitlab.com/my_group/proto"
DbURI: "mysql://user:password@tcp(localhost:3306)/xxx_config"
EnableAssignPort: false
EnableAssignErrcode: true
SvcPortInterval: 5
SvcErrcodeInterval: 1000
SvcGroupInitPortMap:
  base: 9000
  spp: 10000
SvcGroupInitErrcodeMap:
  base: 10000
  spp: 100000
OnceFiles:
  - ".gitignore"
  - "go.mod"
  - ".editorconfig"
  - "README.md"
  - "Dockerfile"
  - "Makefile"
```

## 配置项说明
- `TEMPLATES_BASE_DIR`: 代码模版的指定目录。具体模版要求示例查看[gctl-templates](https://github.com/ml444/gctl-templates)
- `TARGET_BASE_DIR`: 代码生成的指定根目录，默认使用当前系统用户的home目录。
- `MODULE_PREFIX`: go.mod的module名称前缀。
- `DEFAULT_SVC_GROUP`: 有些时候，我们可能长期会在一个分组下开发代码，那么可以设置这个默认分组，这样我们在执行命令时就可以不需要加上指定分组：`-g=my_group`。
- `PROTO_PATHS`: 在你自己定制的模版中如果需要引入一些第三方proto文件时，需要在这里设置第三方proto文件的路径。
- `PROTO_CENTRAL_REPO_PATH`: 如果指定了这个路径说明需要把proto做集中存放处理，生成的proto文件会按照指定的分组放在这个目录下。如果不指定，默认proto文件跟client端代码放在同一目录。
- `DB_URI`: 在代码生成需要自动分配错误码或端口的时候，需要配置数据库的DSN来统一管理错误码或端口的分发。目前支持的数据库类型：`MySQL` 和 `PostgreSQL`。
- `ENABLE_ASSIGN_PORT`: 启用端口自动分配。有时我们可能需要为每个服务设置不同的端口（比如在裸机部署服务的时候），以满足在本地或者同一个机器上同时运行；并且每个服务可能需要多个端口来监听不同的业务或功能。
- `ENABLE_ASSIGN_ERRCODE`: 启用错误码分配。在创建服务的proto文件时，可以启用自动分配错误码范围的功能，这样可以在生成proto时，初步展示第一个错误码，并且标注了该服务的错误码范围。
- `SVC_PORT_INTERVAL`: 端口的步进间隔（默认5），从初始的端口值开始，每个服务的端口区间，默认是每个服务有5个可使用的指定端口（如：[10000～10004]）。
- `SVC_ERRCODE_INTERVAL`: 错误码的步进间隔（默认1000），从初始的端口值开始，每个服务的错误码区间，默认是每个服务有1000个可使用的错误码（如：[100000～100999]）。
- `SVC_GROUP_INIT_PORT_MAP`: 指定各个分组的初始端口，每个服务在该分组内都是通过数据库找到当前最大的服务端口在此基础上增加`SVC_PORT_INTERVAL`，形成本服务可以使用的端口范围。
- `SVC_GROUP_INIT_ERRCODE_MAP`: 指定各个分组的初始错误码，每个服务在该分组内都是通过数据库找到当前最大的服务错误码在此基础上增加`SVC_ERRCODE_INTERVAL`，形成本服务可以使用的错误码范围。
- `ONCE_FILES`: 在重复执行某个服务的生成时，可以指定某些文件不需要重新生成。默认配置：`[".gitignore", "go.mod", ".editorconfig", "README.md", "Dockerfile", "Makefile"]`

## 使用说明
```shell
# 生成proto文件
$ gctl proto -n user -g=my_group
# 生成客户端代码
$ gctl client user -g=my_group
# 生成服务端代码
$ gctl server user -g=my_group
```
如果设置的配置项：`export GCTL_DEFAULT_SERVICE_GROUP="my_group"`，则可以不用输入`-g=my_group`

```shell
# 生成proto文件
$ gctl proto -n user 
# 生成客户端代码
$ gctl client user
# 生成服务端代码
$ gctl server user
```




