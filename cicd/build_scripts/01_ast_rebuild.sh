#!/usr/bin/env bash

CURDIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

REPOSITORY_ROOT="$(realpath "${CURDIR}/../..")"

cd "${REPOSITORY_ROOT}/go/vt/sqlparser"

go run ./visitorgen/main -input=ast.go -output=rewriter.go
