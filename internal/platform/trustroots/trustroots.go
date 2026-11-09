// Package trustroots loads the reviewed Android trust-root set that the
// measurement and context TLS clients use instead of Go's default system
// roots.
//
// On Android the set is the updatable Conscrypt store when that directory
// exists and yields at least one certificate, and otherwise the system store.
// User-installed CAs are never read, whatever the application's network
// security configuration says, and neither are SSL_CERT_FILE, SSL_CERT_DIR or
// any other directory Go's own loader consults. When neither store yields a
// certificate there is no root set: Load returns ErrUnavailable and the caller
// fails TLS-dependent work closed, with no fallback to Go's default roots.
//
// The store paths are fixed in the Android-only build and nothing from a
// task, the host configuration, the trusted local document or the native
// adapters can name another directory or add a certificate.
package trustroots

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Reviewed implementation inputs: the most entries a store directory may
// hold and the most bytes read from one entry. The Android stores hold a few
// hundred files of a few kilobytes each. A directory over the entry bound is
// unusable rather than truncated; a larger entry is skipped.
const (
	MaxEntries    = 1024
	MaxEntryBytes = 64 << 10
)

// ErrUnavailable reports that no store yielded a certificate. It carries no
// path or certificate detail.
var ErrUnavailable = errors.New("trustroots: no trust-root set available")

// Load returns the certificates of the first directory in dirs that yields at
// least one certificate. Each entry of a directory is read as PEM and every
// CERTIFICATE block in it is added, as in the platform's own store format;
// unreadable entries, subdirectories and non-certificate content are skipped.
// Only the named directories are read, never recursively.
func Load(dirs ...string) (*x509.CertPool, error) {
	for _, dir := range dirs {
		if pool, n := loadDir(dir); n > 0 {
			return pool, nil
		}
	}
	return nil, ErrUnavailable
}

// loadDir returns the pool of one directory and how many certificates it
// added. A missing or unreadable directory, or one with more than MaxEntries
// entries, adds none.
func loadDir(dir string) (*x509.CertPool, int) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > MaxEntries {
		return nil, 0
	}
	pool := x509.NewCertPool()
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, ok := readEntry(filepath.Join(dir, e.Name()))
		if !ok {
			continue
		}
		n += appendPEM(pool, data)
	}
	return pool, n
}

// appendPEM adds every parseable CERTIFICATE block of data to pool and
// returns how many it added. Text around the blocks, as the Android store
// files carry, is ignored.
func appendPEM(pool *x509.CertPool, data []byte) int {
	n := 0
	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			return n
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		pool.AddCert(cert)
		n++
	}
}

// readEntry reads at most MaxEntryBytes of one regular file, following a
// symbolic link as the platform stores may use them.
func readEntry(path string) ([]byte, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() || info.Size() > MaxEntryBytes {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(f, MaxEntryBytes+1))
	if err != nil || len(data) > MaxEntryBytes {
		return nil, false
	}
	return data, true
}
