//go:build android

package trustroots

import "crypto/x509"

// The Android store directories, in order of preference. The user-installed
// CA directories (/data/misc/user/<n>/cacerts-added and
// /data/misc/keychain/certs-added) are deliberately absent.
const (
	updatableStore = "/apex/com.android.conscrypt/cacerts"
	systemStore    = "/system/etc/security/cacerts"
)

// Android returns the reviewed Android root set: the updatable Conscrypt
// store when it yields at least one certificate, otherwise the system store,
// otherwise ErrUnavailable.
func Android() (*x509.CertPool, error) { return Load(updatableStore, systemStore) }
