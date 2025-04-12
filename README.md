# GoKitReWi

> Go 语言工具集

## Golang 使用 github 托管 go 类库

### 使用本库

```shell
go get github.com/Re-Wi/GoKitReWi
```

### 升级类库方式

- 使用 go get -u xxx 升级至该主版本号下最新版本；
- 使用 go get xxx@version 升级至指定版本。

> 主版本升级。
> 值得注意的是，使用 go get -u xxx 升级类库版本时，无法跨主版本升级，只能升级至当前主版本下最新小版本；
> v0.x.x 升级至 v1.x.x 是个例外，可以直接使用 go get -u xxx 命令升级。

### 使用本地 go 类库

> 如果本地的 go 类库暂未维护到远端，如何引用本地类库的包呢？
> 在 go.mod 文件中使用 replace 引用本地 go 类库，这个方式有时候更方便于开发。
> 参考：<https://www.zhihu.com/tardis/sogou/art/355318345>

编辑 Go.mod 文件

```mod
module demo-go

go 1.20

require (
 github.com/Re-Wi/GoKitReWi v0.1.1
)

replace (
 github.com/Re-Wi/GoKitReWi => ../GoKitReWi
)

```

## 目录结构

```text
├─databases   连数据库相关
├─handlers    内容处理对象及成员属性
├─helpers     内容处理函数
├─logger      打印日志
├─producer    生产数据
└─tools       可打包为小工具
```

## 参与开发

### 代码编写与测试

```shell
# 依赖包的来源
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.io,direct

# 安装依赖
go mod tidy

# 自行运行测试函数
```

### 提交代码并加版本号

```shell
git add --all
git commit -m "feat: XXXX"
git push
git tag v0.0.0
git push --tags
```

### Go 类库版本规则

go 类库版本的规则：`主版本号.次版本号.修订号`，其中：

- 主版本号：类库进行了不可向下兼容的修改，例如功能重构，这时候主版本号往上追加；
- 次版本号：类库进行了可向下兼容的修改，例如新增功能，这时候次版本号往上追加；
- 修订号：类库进行了可向下兼容的修改（修改的规模更小），例如修复或优化功能，这时候修订好往上追加。

./upgrade.exe create-patch ./FT4222PyTool-V1.0.0/.gitignore ./FT4222PyTool-v1.1.6/.gitignore ./tools/patch.xd -b 8192