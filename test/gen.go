// Package gen hosts go:generate directives for mock generation.
// Run `go generate ./test/...` (or `make generate-mocks`) to regenerate mocks.
package gen

//go:generate go run github.com/vektra/mockery/v2@v2.53.6 --name TranslationLoader --dir ../internal/core/ports --output mocks --outpkg mocks --case underscore
//go:generate go run github.com/vektra/mockery/v2@v2.53.6 --name CacheDriver --dir ../internal/core/ports --output mocks --outpkg mocks --case underscore
//go:generate go run github.com/vektra/mockery/v2@v2.53.6 --name ProductRepository --dir ../internal/core/ports --output mocks --outpkg mocks --case underscore
//go:generate go run github.com/vektra/mockery/v2@v2.53.6 --name DocumentBuilder --dir ../internal/core/ports --output mocks --outpkg mocks --case underscore
