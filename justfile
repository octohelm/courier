[group('meta')]
default:
    @just --list --list-submodules

mod go 'tool/go'

mod example 'internal/example/cmd/example'
