mod example 'internal/example/cmd/example/justfile'

# 列出所有可用命令（无输入）
[group('meta')]
default:
    @just --list

# 运行基础测试
[group('test')]
test path *args:
    go test -count=1 -failfast {{ args }} {{ path }}

# 格式化仓库代码（无输入）
[group('fmt')]
fmt:
    go tool gofumpt -l -w .

# 整理依赖（无输入）
[group('env')]
dep:
    go mod tidy

# 更新依赖版本（无输入）
[group('env')]
update:
    go get -u ./...

# 执行仓库生成
[group('gen')]
gen path:
    go generate {{ path }}
