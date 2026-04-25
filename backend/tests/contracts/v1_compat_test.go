//go:build contracts

// Contract test — current handlers satisfy previous-release OpenAPI v1 spec.
//
// Doctrine refs: Rule 14, Rule 21, Rule 34.
// Plan ADR: docs/adr/0008-api-versioning-policy.md, docs/adr/0013-oapi-codegen-design-first.md.

package contracts_test

import "testing"

// TestV1CompatGoldenSpec runs the previous-release `api/openapi/v1.yaml`
// (committed as `tests/contracts/golden/v1-prev.yaml`) against the current
// handlers via kin-openapi validator. Fails on any breaking change without
// major bump.
func TestV1CompatGoldenSpec(t *testing.T) {
	t.Skip("scaffold — implement when kin-openapi validator wired; first golden snapshot lands at v1.0.0 release")

	// Pseudocode:
	//   prev := loadOpenAPI(t, "golden/v1-prev.yaml")
	//   curr := loadOpenAPI(t, "../../api/openapi/v1.yaml")
	//   for path, op := range prev.Paths {
	//     currOp, ok := curr.Paths[path]
	//     if !ok {
	//       t.Errorf("breaking: path %s removed", path)
	//       continue
	//     }
	//     // every prev request param must exist in curr
	//     // every prev required field must still be required
	//     // every prev response status must still be present
	//     // ... etc
	//   }
}
