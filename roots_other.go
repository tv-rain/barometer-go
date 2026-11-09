//go:build !android

package barometer

import "crypto/x509"

// Elsewhere the seam leaves the roots nil, which selects Go's system roots.
var platformRoots func() (*x509.CertPool, error)
