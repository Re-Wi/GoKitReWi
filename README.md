# GoKitReWi

Go语言工具集

## Golang 使用github托管go类库

```shell
git add --all
git commit -m "Initial Commit"
git push
git tag v0.0.0 
git push --tags
```

## Go类库版本规则

go类库版本的规则：`主版本号.次版本号.修订号`，其中：

- 主版本号：类库进行了不可向下兼容的修改，例如功能重构，这时候主版本号往上追加；
- 次版本号：类库进行了可向下兼容的修改，例如新增功能，这时候次版本号往上追加；
- 修订号：类库进行了可向下兼容的修改（修改的规模更小），例如修复或优化功能，这时候修订好往上追加。

## 升级类库方式

- 使用go get -u xxx升级至该主版本号下最新版本；
- 使用go get xxx@version升级至指定版本。

> 主版本升级。
> 值得注意的是，使用go get -u xxx升级类库版本时，无法跨主版本升级，只能升级至当前主版本下最新小版本；
> v0.x.x 升级至v1.x.x是个例外，可以直接使用go get -u xxx命令升级。

## 使用本地go类库

> 如果本地的go类库暂未维护到远端，如何引用本地类库的包呢？
> 在go.mod文件中使用replace引用本地go类库，这个方式有时候更方便于开发。
> 参考：https://www.zhihu.com/tardis/sogou/art/355318345