[group('meta')]
default:
    @just --list --list-submodules

[group: 'toolchain']
mod go 'tool/go'

[group: 'api-server']
mod example 'internal/example/cmd/example'
