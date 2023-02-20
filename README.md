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