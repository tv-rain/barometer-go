package barometermobile

// Native adapter callback interfaces. Barometer-owned Kotlin and Swift code
// implements them; they are an adapter boundary, not a host extension point.
//
// Every blocking method receives its remaining deadline as timeoutMillis and
// must return by then. The facade also stops waiting at the deadline and
// treats the call as failed; a call it abandoned must not take effect after a
// later clear or close, and an uncertain outcome is reported as an error,
// never as success. Callbacks run on Go-created threads: native code must not
// block them on the UI thread or call back into the facade from inside a
// callback. Any error or unknown value is unsatisfied or unavailable.

// Records is the application-scoped, non-synchronizing byte record store of
// one storage scope, excluded from backup and sync. A key has the form
// "<class>-<64 lowercase hex digits>", for example "3-ab…", with class 1 to 4;
// it is safe to use as a file name and names no issuer, task or other value.
type Records interface {
	// Get returns the stored value, or NewAbsentRecordValue() when the key is
	// known to be absent. A nil value is unavailable, like an error: Swift
	// cannot return nil without throwing.
	Get(key string, timeoutMillis int64) (*RecordValue, error)
	// Put atomically stores value for key. On error either the earlier value
	// or value is durable, never a mixture.
	Put(key string, value []byte, timeoutMillis int64) error
	// Delete atomically removes key; deleting an absent key succeeds.
	Delete(key string, timeoutMillis int64) error
	// List returns every stored key of class, in any order.
	List(class int64, timeoutMillis int64) (*StringList, error)
}

// LinkageKeyStore derives HMAC-SHA-256 inside protected, non-exporting,
// non-synchronizing key storage, creating the key on first use. Delete is
// idempotent and creates no replacement.
type LinkageKeyStore interface {
	// DeriveHMACSHA256 returns exactly 32 bytes.
	DeriveHMACSHA256(message []byte, timeoutMillis int64) ([]byte, error)
	Delete(timeoutMillis int64) error
}

// Clock reports trusted wall-clock time in Unix milliseconds. Unavailable
// time is an error, never a fallback to the system clock; a value that is not
// positive is unavailable.
type Clock interface {
	NowUnixMillis(timeoutMillis int64) (int64, error)
}

// Network reports coarse access state and one-shot access-network changes.
type Network interface {
	Access(timeoutMillis int64) (*AccessState, error)
	// Watch arranges for observer.Changed to be called once, from any thread,
	// when the access-network generation is no longer generation,
	// immediately if it already differs. It must return promptly. Changed
	// calls Access and waits for it, to learn whether the change is a
	// suspension, so native code must not call Changed while holding a lock
	// that Access needs.
	Watch(generation int64, observer NetworkObserver) (int64, error)
	// StopWatch cancels a watch. It is idempotent and does not wait for a
	// Changed call in progress.
	StopWatch(watchID int64)
}

// NetworkObserver is implemented by the facade and passed to Network.Watch.
type NetworkObserver interface {
	Changed()
}

// Consent reports consent status with no task value: "valid", "absent" or
// "ended". Any other value is unavailable.
type Consent interface {
	Status(timeoutMillis int64) (string, error)
}

// Device reports coarse battery and charging state.
type Device interface {
	State(timeoutMillis int64) (*DeviceState, error)
}

// TimeZone reports the device's IANA time-zone identifier.
type TimeZone interface {
	Identifier(timeoutMillis int64) (string, error)
}

// SystemResolver performs one ordinary platform lookup of host for network
// "ip", "ip4" or "ip6" and returns the addresses in canonical text form.
type SystemResolver interface {
	LookupNetIP(network string, host string, timeoutMillis int64) (*StringList, error)
}

// Adapters are the native adapters passed to NewClient. Records, Clock and
// Network are required. Consent, Device, TimeZone and Resolver are optional: a
// task that needs a missing one never runs. LinkageKeys is optional: without
// it only result linkage is disabled.
type Adapters struct {
	Records     Records
	Clock       Clock
	Network     Network
	Consent     Consent
	Device      Device
	TimeZone    TimeZone
	Resolver    SystemResolver
	LinkageKeys LinkageKeyStore
}

// RecordValue is the result of Records.Get: one stored record value, or the
// known absence of one.
type RecordValue struct {
	value  []byte
	absent bool
}

// NewRecordValue copies value.
func NewRecordValue(value []byte) *RecordValue {
	return &RecordValue{value: append([]byte{}, value...)}
}

// NewAbsentRecordValue reports that the key is known to be absent.
func NewAbsentRecordValue() *RecordValue { return &RecordValue{absent: true} }

// maxListEntries bounds a StringList; a longer list is unavailable.
const maxListEntries = 4096

// StringList is a bounded list of strings built by native code.
type StringList struct {
	items    []string
	overflow bool
}

// NewStringList returns an empty list.
func NewStringList() *StringList { return &StringList{} }

// Add appends s. Adding more than 4,096 entries makes the list unavailable.
func (l *StringList) Add(s string) {
	if len(l.items) >= maxListEntries {
		l.overflow = true
		return
	}
	l.items = append(l.items, s)
}

// AccessState is one coarse access-context reading. It has no field for an
// interface name, address, DNS server, Wi-Fi or cell identifier, SIM,
// subscriber, carrier or location.
type AccessState struct {
	generation                                              int64
	networkType, vpn, ipv4, ipv6, validated, cpo, suspended string
}

// NewAccessState builds a reading. generation identifies the access network
// and changes whenever it changes; zero means unknown. networkType is "wifi",
// "cellular", "ethernet" or "other"; vpn is "detected" or "not_detected"; ipv4,
// ipv6, validated and captivePortal are "yes" or "no"; suspended is
// "suspended" when mobile lifecycle or operating-system policy has suspended
// the application's execution and "active" when it has not. Any other value,
// and a negative generation, is unknown; an unknown suspension state does not
// block execution. Entering or leaving suspension must advance the generation.
func NewAccessState(generation int64, networkType string, vpn string, ipv4 string, ipv6 string, validated string, captivePortal string, suspended string) *AccessState {
	return &AccessState{generation: generation, networkType: networkType, vpn: vpn, ipv4: ipv4, ipv6: ipv6, validated: validated, cpo: captivePortal, suspended: suspended}
}

// DeviceState is a coarse device snapshot.
type DeviceState struct {
	batteryPercent                        int64
	batteryKnown, charging, chargingKnown bool
}

// NewDeviceState builds a snapshot. A battery percentage outside 0 to 100 is
// unknown.
func NewDeviceState(batteryPercent int64, batteryKnown bool, charging bool, chargingKnown bool) *DeviceState {
	return &DeviceState{batteryPercent: batteryPercent, batteryKnown: batteryKnown, charging: charging, chargingKnown: chargingKnown}
}
