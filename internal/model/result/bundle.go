package result

// RotatingLinkage is a detached projection supplied by the future trusted
// private-state owner. Token is the derived HMAC output, never a secret or a
// host identifier. Its provenance and epoch lifecycle are runtime obligations.
// There is deliberately no stable mode, previous token or continuity mapping.
type RotatingLinkage struct {
	Scope string
	Epoch int64
	Token [32]byte
}

type Bundle struct {
	SchemaVersion, BundleID string
	Linkage                 *RotatingLinkage
	Results                 []Result
}

// Only the validated bundle boundary may serialize these detached projections.
func (Bundle) MarshalJSON() ([]byte, error)          { return nil, ErrEncoding }
func (*Bundle) UnmarshalJSON([]byte) error           { return ErrEncoding }
func (RotatingLinkage) MarshalJSON() ([]byte, error) { return nil, ErrEncoding }
func (*RotatingLinkage) UnmarshalJSON([]byte) error  { return ErrEncoding }
