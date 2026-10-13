//go:build barometer_mobile_smoke

package barometer

// MobileSmoke echoes a synthetic value solely to exercise mobile binding builds.
// It is excluded from normal builds and is not a production API.
func MobileSmoke(value string) string { return value }
