// Unit tests for manifest validation. Lives alongside manifest.go so the
// in-package contract is reviewable as a single artefact.
//
// Doctrine refs: Rule 44 (testability), Rule 70 (Pack manifests itself).

package packs

import "testing"

func TestValidateBasic_Happy(t *testing.T) {
	m := PackManifest{
		ID:                  "modbus_tcp",
		Kind:                KindProtocol,
		Version:             "1.0.0",
		MinCoreVersion:      "1.0.0",
		PackContractVersion: "1.0.0",
		Author:              "Macena GreenMetrics",
		LicenseSPDX:         "Proprietary",
		Capabilities:        []string{"protocol.modbus.tcp"},
	}
	if err := m.ValidateBasic(); err != nil {
		t.Fatalf("happy path failed: %v", err)
	}
}

func TestValidateBasic_RejectsBadID(t *testing.T) {
	m := PackManifest{
		ID:                  "Modbus TCP",
		Kind:                KindProtocol,
		Version:             "1.0.0",
		MinCoreVersion:      "1.0.0",
		PackContractVersion: "1.0.0",
		Author:              "Macena",
		LicenseSPDX:         "Proprietary",
		Capabilities:        []string{"x"},
	}
	if err := m.ValidateBasic(); err == nil {
		t.Fatal("uppercase + space id should be rejected")
	}
}

func TestValidateBasic_RejectsBadVersion(t *testing.T) {
	m := PackManifest{
		ID:                  "modbus_tcp",
		Kind:                KindProtocol,
		Version:             "v1.0.0",
		MinCoreVersion:      "1.0.0",
		PackContractVersion: "1.0.0",
		Author:              "Macena",
		LicenseSPDX:         "Proprietary",
		Capabilities:        []string{"x"},
	}
	if err := m.ValidateBasic(); err == nil {
		t.Fatal("v-prefix is not SemVer; should be rejected")
	}
}

func TestValidateBasic_RejectsUnknownKind(t *testing.T) {
	m := PackManifest{
		ID:                  "weird",
		Kind:                PackKind("dashboard"),
		Version:             "1.0.0",
		MinCoreVersion:      "1.0.0",
		PackContractVersion: "1.0.0",
		Author:              "Macena",
		LicenseSPDX:         "Proprietary",
		Capabilities:        []string{"x"},
	}
	if err := m.ValidateBasic(); err == nil {
		t.Fatal("unknown kind should be rejected")
	}
}

func TestValidateBasic_RequiresLicense(t *testing.T) {
	m := PackManifest{
		ID:                  "modbus_tcp",
		Kind:                KindProtocol,
		Version:             "1.0.0",
		MinCoreVersion:      "1.0.0",
		PackContractVersion: "1.0.0",
		Author:              "Macena",
		LicenseSPDX:         "",
		Capabilities:        []string{"x"},
	}
	if err := m.ValidateBasic(); err == nil {
		t.Fatal("empty license_spdx should be rejected")
	}
}

func TestValidateBasic_RequiresAtLeastOneCapability(t *testing.T) {
	m := PackManifest{
		ID:                  "modbus_tcp",
		Kind:                KindProtocol,
		Version:             "1.0.0",
		MinCoreVersion:      "1.0.0",
		PackContractVersion: "1.0.0",
		Author:              "Macena",
		LicenseSPDX:         "Proprietary",
		Capabilities:        nil,
	}
	if err := m.ValidateBasic(); err == nil {
		t.Fatal("empty capabilities should be rejected")
	}
}

// SemVer corner cases — pre-release + build metadata.

func TestValidateBasic_AcceptsSemVerPreRelease(t *testing.T) {
	m := PackManifest{
		ID:                  "modbus_tcp",
		Kind:                KindProtocol,
		Version:             "1.0.0-rc.1",
		MinCoreVersion:      "1.0.0",
		PackContractVersion: "1.0.0",
		Author:              "Macena",
		LicenseSPDX:         "Proprietary",
		Capabilities:        []string{"x"},
	}
	if err := m.ValidateBasic(); err != nil {
		t.Fatalf("pre-release version should be valid SemVer: %v", err)
	}
}

func TestValidateBasic_AcceptsSemVerBuildMetadata(t *testing.T) {
	m := PackManifest{
		ID:                  "modbus_tcp",
		Kind:                KindProtocol,
		Version:             "1.0.0+build.123",
		MinCoreVersion:      "1.0.0",
		PackContractVersion: "1.0.0",
		Author:              "Macena",
		LicenseSPDX:         "Proprietary",
		Capabilities:        []string{"x"},
	}
	if err := m.ValidateBasic(); err != nil {
		t.Fatalf("build-metadata version should be valid SemVer: %v", err)
	}
}
