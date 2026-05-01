package services

import (
	"testing"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

func TestPackCount(t *testing.T) {
	region, factor, report := PackCount()
	if region != 1 {
		t.Errorf("region count: want 1 (region-it), got %d", region)
	}
	if factor != 4 {
		t.Errorf("factor count: want 4 (ispra+gse+terna+aib), got %d", factor)
	}
	if report != 7 {
		t.Errorf("report count: want 7 (monthly_consumption+co2_footprint+esrs_e1+piano_5_0+audit_dlgs102+tee+conto_termico), got %d", report)
	}
}

// TestReportPacksAllRegistered verifies the 7 Italian-flagship Report Packs
// are present under their declared ReportType. This is the runtime invariant
// that the dispatch path (Phase F) depends on.
func TestReportPacksAllRegistered(t *testing.T) {
	want := []reporting.ReportType{
		"monthly_consumption",
		"co2_footprint",
		"esrs_e1",
		"piano_5_0",
		"audit_dlgs102",
		"tee",
		"conto_termico",
	}
	registry := ReportPacks()
	for _, rt := range want {
		pack, ok := registry[rt]
		if !ok {
			t.Errorf("Report Pack %q not registered", rt)
			continue
		}
		if pack.Type() != rt {
			t.Errorf("Pack registered under %q reports Type() = %q (mismatch)", rt, pack.Type())
		}
		if pack.Version() == "" {
			t.Errorf("Pack %q has empty Version()", rt)
		}
	}
}

// TestFactorPacksAllRegistered verifies the 4 Italian-flagship Factor Packs
// are present under their declared Name.
func TestFactorPacksAllRegistered(t *testing.T) {
	want := []string{"ispra", "gse", "terna", "aib"}
	registry := FactorPacks()
	for _, name := range want {
		pack, ok := registry[name]
		if !ok {
			t.Errorf("Factor Pack %q not registered", name)
			continue
		}
		if pack.Name() != name {
			t.Errorf("Pack registered under %q reports Name() = %q (mismatch)", name, pack.Name())
		}
	}
}

// TestRegionPacksAllRegistered verifies the 1 Italian-flagship Region Pack.
func TestRegionPacksAllRegistered(t *testing.T) {
	registry := RegionPacks()
	pack, ok := registry["it"]
	if !ok {
		t.Fatal("Region Pack \"it\" not registered")
	}
	if pack.Code() != "it" {
		t.Errorf("Pack registered under \"it\" reports Code() = %q (mismatch)", pack.Code())
	}
}

// TestRegistryReturnsDefensiveCopy — mutating the returned map MUST NOT
// affect subsequent calls.
func TestRegistryReturnsDefensiveCopy(t *testing.T) {
	a := ReportPacks()
	b := ReportPacks()
	if len(a) != len(b) {
		t.Fatal("two calls returned different counts")
	}
	delete(a, "monthly_consumption")
	if _, ok := b["monthly_consumption"]; !ok {
		t.Fatal("mutation in one map affected the other (registry is not returning a defensive copy)")
	}
	c := ReportPacks()
	if _, ok := c["monthly_consumption"]; !ok {
		t.Fatal("mutation in one map affected a subsequent call")
	}
}

// TestNoDuplicateReportTypes — the registry is a one-to-one map; the
// constructor-time loop would silently overwrite a duplicate, so we verify
// uniqueness explicitly.
func TestNoDuplicateReportTypes(t *testing.T) {
	if len(ReportPacks()) != 7 {
		t.Errorf("expected exactly 7 unique ReportTypes, got %d (silent overwrite suspect)", len(ReportPacks()))
	}
}
