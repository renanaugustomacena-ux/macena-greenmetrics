// Manifest schema and validation for Packs.
//
// Doctrine refs: Rules 70 (Pack manifests itself), 75 (capabilities declared),
// 83 (no recursive Pack discovery — manifest is the only declaration surface).
// Charter ref: §4 (Pack Contract — the formal seam).
//
// Every Pack ships a manifest.yaml at packs/<kind>/<id>/manifest.yaml. The
// schema lives at docs/contracts/pack-manifest.schema.json and is the source
// of truth; the Go struct below mirrors it. CI validates schema/struct
// drift via tests/packs/manifest_validation_test.go (Sprint S5 deliverable).

package packs

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// PackManifest is the parsed shape of a per-Pack manifest.yaml.
//
// Fields with tag `validate:"required,..."` are validated by go-playground/validator
// (ADR-0012). The custom `oneof_pack_id` validator enforces Rule 70's id
// canonical pattern.
type PackManifest struct {
	// Identity (required).
	ID                  string   `json:"id" yaml:"id" validate:"required,oneof_pack_id"`
	Kind                PackKind `json:"kind" yaml:"kind" validate:"required,oneof=protocol factor report identity region"`
	Version             string   `json:"version" yaml:"version" validate:"required,semver"`
	MinCoreVersion      string   `json:"min_core_version" yaml:"min_core_version" validate:"required,semver"`
	PackContractVersion string   `json:"pack_contract_version" yaml:"pack_contract_version" validate:"required,semver"`

	// Provenance (required).
	Author      string `json:"author" yaml:"author" validate:"required"`
	LicenseSPDX string `json:"license_spdx" yaml:"license_spdx" validate:"required"`

	// Behaviour (required).
	Capabilities []string `json:"capabilities" yaml:"capabilities" validate:"required,min=1,dive,required"`

	// Optional fields.
	Notes        string   `json:"notes,omitempty" yaml:"notes,omitempty"`
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Homepage     string   `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Repository   string   `json:"repository,omitempty" yaml:"repository,omitempty"`
}

// idPattern matches Pack ids: lowercase alphanumeric, underscores, hyphens.
// Length 3..64. Examples: modbus_tcp, ispra, esrs_e1, region-it.
var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{2,63}$`)

// semverPattern matches a permissive subset of SemVer 2.0.0. Strict-validation
// upgrades happen via the validator/v10 `semver` rule which we register in
// the Pack-loader bootstrap.
var semverPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

// ValidateBasic runs the manifest-self checks that don't require the validator
// package. It is the cheapest possible reject — used at boot before the
// full validator pass to surface obvious malformation.
//
// Returns nil when the manifest is at least minimally well-shaped. The
// loader runs this and then runs the full validator pass; both must pass.
func (m PackManifest) ValidateBasic() error {
	if !idPattern.MatchString(m.ID) {
		return fmt.Errorf("manifest id %q does not match canonical pattern %s", m.ID, idPattern)
	}
	if !semverPattern.MatchString(m.Version) {
		return fmt.Errorf("manifest version %q is not SemVer", m.Version)
	}
	if !semverPattern.MatchString(m.MinCoreVersion) {
		return fmt.Errorf("manifest min_core_version %q is not SemVer", m.MinCoreVersion)
	}
	if !semverPattern.MatchString(m.PackContractVersion) {
		return fmt.Errorf("manifest pack_contract_version %q is not SemVer", m.PackContractVersion)
	}
	switch m.Kind {
	case KindProtocol, KindFactor, KindReport, KindIdentity, KindRegion:
	default:
		return fmt.Errorf("manifest kind %q is not a recognised PackKind", m.Kind)
	}
	if strings.TrimSpace(m.Author) == "" {
		return errors.New("manifest author empty")
	}
	if strings.TrimSpace(m.LicenseSPDX) == "" {
		return errors.New("manifest license_spdx empty — Pack license is required")
	}
	if len(m.Capabilities) == 0 {
		return errors.New("manifest capabilities empty — at least one capability is required (Rule 75)")
	}
	for _, c := range m.Capabilities {
		if strings.TrimSpace(c) == "" {
			return errors.New("manifest capability empty string")
		}
	}
	return nil
}
