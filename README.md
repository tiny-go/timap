# Timap

[![Report Card][report-badge]][report-link]
[![GoCover][cover-badge]][cover-link]

# Different map implementations

- `Timap` is a wrapper over `sync.Map` that allows to set life time for each key-value pair.
- `CtxMap` allows storing key/value pairs with context to be deleted once the context is cancelled or its deadline is exceeded. 

[report-badge]: https://goreportcard.com/badge/github.com/tiny-go/timap
[report-link]: https://goreportcard.com/report/github.com/tiny-go/timap
[cover-badge]: https://gocover.io/_badge/github.com/tiny-go/timap
[cover-link]: https://gocover.io/github.com/tiny-go/timap
