

# StackQL Parser

This is the `stackql` parser, a forked descendent of [vitess](https://github.com/vitessio/vitess); we are deeply grateful to and fully acknowledge this work.

There are elements of the original work that are not required, but may take some time to excise.


## Rebuilding parser


```
make -C go/vt/sqlparser
```


After changes to the ast:

```
cd go/vt/sqlparser

go run ./visitorgen/main -input=ast.go -output=rewriter.go
```


## License

Unless otherwise noted, source files are distributed
under the Apache Version 2.0 license found in the LICENSE file.

