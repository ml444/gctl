# GCTL
[中文](README_CN.md)

A code generation and checking tool for Go microservices

## Install protoc
> `gctl` requires `protoc`,so please install [protoc](https://github.com/protocolbuffers/protobuf/releases) before you start using it. 
1. Install proto3 from binary
   Download the latest release from [protoc](https://github.com/protocolbuffers/protobuf/releases)
2. Install proto3 from source macOS only.
   ```shell
   > brew install autoconf automake libtool
   > git clone https://github.com/google/protobuf
   > ./autogen.sh ; ./configure ; make ; make install
   ```
3. Install protoc from source.
   ```shell
   sudo apt-get install -y git autoconf automake libtool curl make g++ unzip
   git clone https://github.com/google/protobuf.git
   cd protobuf/
   ./autogen.sh
   ./configure
   make
   make check
   sudo make install
   sudo ldconfig # refresh shared library cache.
   ```
### Install protoc plugins
```shell
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/ml444/gkit/cmd/protoc-gen-go-field@latest
	go install github.com/ml444/gkit/cmd/protoc-gen-go-http@latest
	go install github.com/ml444/gkit/cmd/protoc-gen-go-gorm@latest
	go install github.com/ml444/gkit/cmd/protoc-gen-go-errcode@latest
	go install github.com/ml444/gkit/cmd/protoc-gen-go-validate@latest
	go install github.com/ml444/gctl@latest
```
**OpenAPI Swagger options**:
- openapi.yaml: `go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest` -> `--openapi_out=.`
- swagger.json(gctl default): `go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest` -> `--openapiv2_out=.`

# See also
> https://github.com/grpc/grpc-go/tree/master/examples


## Usage


## 配置
1. **Configuration by environment variables**
    ```shell
    # ---required configurations---
    #export GCTL_TEMPLATES_BASE_DIR="/your_path/github.com/ml444/gctl-templates/centralized_templates"
    export GCTL_TEMPLATES_BASE_DIR="/your_path/github.com/ml444/gctl-templates/separation_templates"
    export GCTL_PROTO_PATHS="/your_path/github.com/ml444/gctl-templates/protofiles"
    export GCTL_MODULE_PREFIX="my.gitlab.com"
    ```
    ```shell
    # ---optional configurations---
    export GCTL_TARGET_BASE_DIR="/your/target/root/path"
    export GCTL_PROTO_CENTRAL_REPO_PATH="/your_path/my.gitlab.com/my_group/proto"
    export GCTL_ONCE_FILES=".gitignore go.mod .editorconfig README.md Makefile Dockerfile"
    export GCTL_DEFAULT_SERVICE_GROUP="my_group"
    export GCTL_DB_DSN="mysql://user:password@tcp(localhost:3306)/my_config"
    export GCTL_ENABLE_ASSIGN_PORT=false
    export GCTL_ENABLE_ASSIGN_ERRCODE=true
    export GCTL_SVC_PORT_INTERVAL=5
    export GCTL_SVC_ERRCODE_INTERVAL=1000
    export GCTL_SVC_GROUP_INIT_PORT_MAP={"base": 9000, "biz": 10000}
    export GCTL_SVC_GROUP_INIT_ERRCODE_MAP={"base": 10000, "biz": 100000}
    ```

2. **Configuration by yaml file**

   Create a configuration file in the user root directory: for example in Linux `~/.gctl_config.yaml`.

   The following example contains all configuration items.
   Except for the required configuration, the rest can be deleted as needed.
    ```yaml
    #TemplatesBaseDir: "/your_path/github.com/ml444/gctl-templates/centralized_templates"
    TemplatesBaseDir: "/your_path/github.com/ml444/gctl-templates/separation_templates"
    TargetBaseDir: "/your/target/root/path"
    ModulePrefix: "my.gitlab.com"
    DefaultServiceGroup: "my_group"
    ProtoPaths: 
      - "/your_path/github.com/ml444/gctl-templates/protofiles"
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

## Description of configuration options
- `TEMPLATES_BASE_DIR`: The specified directory for the code templates. See [gctl-templates](https://github.com/ml444/gctl-templates) for a sample template requirement
- `TARGET_BASE_DIR`: The specified root directory for code generation, the home directory of the current system user is used by default.
- `MODULE_PREFIX`: The module name prefix for go.mod.
- `DEFAULT_SVC_GROUP`: Sometimes we may develop code under a group for a long time, so we can set this default group so that we can execute commands without adding the specified group: `-g=my_group`.
- `PROTO_PATHS`: If you need to introduce some third-party proto files in your own custom template, you need to set the path of the third-party proto files here.
- `PROTO_CENTRAL_REPO_PATH`: If this path is specified, it means that you need to centralize the proto files, and the generated proto files will be placed in this directory according to the specified group. If not specified, the default proto files will be placed in the same directory as the client code.
- `DB_URI`: When the code generation needs to automatically assign error codes or ports, you need to configure the DSN of the database to unify the distribution of error codes or ports. Currently supported database types: `MySQL` and `PostgreSQL`.
- `ENABLE_ASSIGN_PORT`: Enables automatic port assignment. Sometimes we may need to set different ports for each service (e.g. when deploying services on bare metal) to accommodate running locally or on the same machine at the same time; and each service may need multiple ports to listen to different services or functions.
- `ENABLE_ASSIGN_ERRCODE`: Enables error code assignment. When creating the proto file for a service, you can enable the automatic assignment of error code ranges, so that the first error code is initially displayed when the proto is generated and the error code range for the service is marked.
- `SVC_PORT_INTERVAL`: Step interval of ports (default 5), starting from the initial port value, the port interval for each service, the default is 5 specified ports available for each service (e.g., [10000 to 10004]).
- `SVC_ERRCODE_INTERVAL`: Step interval of error codes (default 1000), starting from the initial port value, the error code interval for each service, the default is 1000 available error codes for each service (e.g. [100000 to 100999]).
- `SVC_GROUP_INIT_PORT_MAP`: Specify the initial port of each group, each service within the grouping is to find the current maximum service port through the database to add `SVC_PORT_INTERVAL` on top of this to form a range of ports that can be used by this service.
- `SVC_GROUP_INIT_ERRCODE_MAP`: Specify the initial error code of each group, each service in the group is to find the current maximum service error code through the database and add `SVC_ERRCODE_INTERVAL` on top of it to form the range of error codes that can be used by this service.
- `ONCE_FILES`: When repeatedly performing the generation of a service, you can specify that certain files do not need to be regenerated. Default configuration: `[".gitignore", "go.mod", ".editorconfig", "README.md", "Dockerfile", "Makefile"]`

## Instructions for use
```shell
# Generate proto file
$ gctl proto -n user -g=my_group
# Generate client-side code
$ gctl client user -g=my_group
# Generate server-side code
$ gctl server user -g=my_group
```
If you set the configuration option: `export GCTL_DEFAULT_SERVICE_GROUP="my_group"`, then you can leave out the `-g=my_group`
```shell
# Generate proto file
$ gctl proto -n user 
# Generate client-side code
$ gctl client user
# Generate server-side code
$ gctl server user
```




