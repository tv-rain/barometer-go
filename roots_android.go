//go:build android

package barometer

import "github.com/tv-rain/barometer-go/internal/platform/trustroots"

// On Android the seam's trust roots are the reviewed platform root set
// instead of Go's default system roots, which add user-installed CAs and
// never read the updatable store.
var platformRoots = trustroots.Android
