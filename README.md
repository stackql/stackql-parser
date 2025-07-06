

# StackQL Parser

This is the `stackql` parser, a forked descendent of [vitess](https://github.com/vitessio/vitess); we are deeply grateful to and fully acknowledge this work.

There are elements of the original work that are not required, but may take some time to excise.


## Rebuilding parser


```bash
make -C go/vt/sqlparser
```


After changes to the ast:

```bash
cicd/build_scripts/01_ast_rebuild.sh
```


## License

Unless otherwise noted, source files are distributed
under the Apache Version 2.0 license found in the LICENSE file.

