//go:build cgo

package crypto

// CGOEnabled is true when this package is built with CGO support.
// Use this flag at runtime to detect whether native C-backed PQ primitives are available.
const CGOEnabled = true
