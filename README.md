# lungo-go

The Go support library of [lungo](https://github.com/machinefabric/lungo): the wire format
and the calls into lungo's runtime that the Go packages lungo generates build on. A generated
package imports it; you do not use it directly.

```sh
go get github.com/machinefabric/lungo-go@v<lungo version>
```

Each release is tagged on the `dist` branch, whose tree carries the prebuilt runtime of every
platform cgo links; `main` holds the source. A generated package requires exactly the release
of lungo that generated it. Documentation: <https://lungo.machinefabric.com/docs>.

Licensed under the [Apache License, Version 2.0](LICENSE).
