package trustroots

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// synthCA is a synthetic CA generated when the test runs. No platform or
// real certificate is used.
type synthCA struct {
	cert *x509.Certificate
	pem  []byte
}

func newCA(t *testing.T, name string) synthCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: name},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return synthCA{cert: cert, pem: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})}
}

// storeFile is the Android store file form: the PEM block followed by a text
// dump of the certificate.
func (c synthCA) storeFile() []byte {
	return append(append([]byte{}, c.pem...), []byte("Certificate:\n    Data:\n        Version: 3 (0x2)\n")...)
}

func trusted(pool *x509.CertPool, c synthCA) bool {
	_, err := c.cert.Verify(x509.VerifyOptions{Roots: pool})
	return err == nil
}

func store(t *testing.T, files map[string][]byte) string {
	t.Helper()
	dir := t.TempDir()
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// The updatable store is used when it yields a certificate, and then the
// system store is not read; otherwise the system store is used.
func TestLoadPrefersTheUpdatableStore(t *testing.T) {
	updated, system := newCA(t, "updatable"), newCA(t, "system")
	upd := store(t, map[string][]byte{"a.0": updated.storeFile()})
	sys := store(t, map[string][]byte{"b.0": system.storeFile()})

	pool, err := Load(upd, sys)
	if err != nil || !trusted(pool, updated) || trusted(pool, system) {
		t.Fatal("the updatable store was not the root set")
	}
	for name, first := range map[string]string{
		"missing":       filepath.Join(t.TempDir(), "absent"),
		"empty":         t.TempDir(),
		"no CA content": store(t, map[string][]byte{"x.0": []byte("not a certificate\n")}),
		"a file":        filepath.Join(upd, "a.0"),
	} {
		pool, err := Load(first, sys)
		if err != nil || !trusted(pool, system) || trusted(pool, updated) {
			t.Fatalf("%s updatable store: the system store was not the root set", name)
		}
	}
}

// With no store yielding a certificate there is no root set, and a
// synthetic CA outside the stores is never trusted.
func TestLoadFailsClosed(t *testing.T) {
	if pool, err := Load(t.TempDir(), filepath.Join(t.TempDir(), "absent")); pool != nil || !errors.Is(err, ErrUnavailable) {
		t.Fatal("no store must be unavailable")
	}
	if pool, err := Load(); pool != nil || !errors.Is(err, ErrUnavailable) {
		t.Fatal("no directory must be unavailable")
	}
	in, outside := newCA(t, "in"), newCA(t, "outside")
	pool, err := Load(store(t, map[string][]byte{"a.0": in.storeFile()}))
	if err != nil || !trusted(pool, in) || trusted(pool, outside) {
		t.Fatal("a CA outside the store was trusted")
	}
}

// Entries are read one level deep and within their bounds: subdirectories,
// oversized entries, malformed blocks and non-certificate blocks add
// nothing, and a directory over the entry bound is unusable.
func TestLoadBounds(t *testing.T) {
	ok, nested, big, other := newCA(t, "ok"), newCA(t, "nested"), newCA(t, "big"), newCA(t, "other")
	dir := store(t, map[string][]byte{
		"ok.0":    ok.storeFile(),
		"big.0":   append(big.storeFile(), make([]byte, MaxEntryBytes)...),
		"key.0":   pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{1}}),
		"bad.0":   pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte{1, 2, 3}}),
		"other.0": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Headers: map[string]string{"Proc-Type": "4,ENCRYPTED"}, Bytes: other.cert.Raw}),
	})
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "n.0"), nested.storeFile(), 0o600); err != nil {
		t.Fatal(err)
	}
	pool, err := Load(dir)
	if err != nil || !trusted(pool, ok) || trusted(pool, nested) || trusted(pool, big) || trusted(pool, other) {
		t.Fatal("an entry outside the bounds was loaded")
	}

	crowded := t.TempDir()
	for i := 0; i <= MaxEntries; i++ {
		name := filepath.Join(crowded, "e"+strconv.Itoa(i)+".0")
		if err := os.WriteFile(name, ok.storeFile(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if pool, err := Load(crowded); pool != nil || !errors.Is(err, ErrUnavailable) {
		t.Fatal("a directory over the entry bound was used")
	}
}

// The Android build names exactly the updatable and system stores, in that
// order, and nothing in the package names a user CA directory or Go's
// certificate environment variables. android.go is not compiled on the
// host, so its source is checked.
func TestAndroidStoresExcludeUserCAs(t *testing.T) {
	src, err := os.ReadFile("android.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	if !strings.HasPrefix(s, "//go:build android\n") ||
		!strings.Contains(s, `updatableStore = "/apex/com.android.conscrypt/cacerts"`) ||
		!strings.Contains(s, `systemStore    = "/system/etc/security/cacerts"`) ||
		!strings.Contains(s, "return Load(updatableStore, systemStore)") ||
		strings.Count(s, `"/`) != 2 {
		t.Fatal("the Android store list changed")
	}
	for _, name := range []string{"android.go", "trustroots.go"} {
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		code := stripComments(string(src))
		for _, banned := range []string{"cacerts-added", "certs-added", "/data/", "SSL_CERT", "Getenv", "SystemCertPool"} {
			if strings.Contains(code, banned) {
				t.Fatalf("%s names %q", name, banned)
			}
		}
	}
}

func stripComments(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		if i := strings.Index(line, "//"); i >= 0 && !strings.Contains(line[:i], `"`) {
			line = line[:i]
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}
