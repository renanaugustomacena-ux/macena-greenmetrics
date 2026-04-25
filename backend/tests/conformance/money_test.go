//go:build conformance

// Conformance — Money fields in API responses are (amount_cents int64, currency ISO-4217 string),
// never float. Verified at the contract layer (api/openapi/v1.yaml) + AST scan (tests/static/).

package conformance_test

import "testing"

func TestMoneyShapeInOpenAPI(t *testing.T) {
	t.Skip("scaffold — load api/openapi/v1.yaml and assert every Money schema is { amount_cents:integer, currency:string-3 }")

	// Pseudocode:
	// spec := loadOpenAPI(t, "../../../api/openapi/v1.yaml")
	// for path, op := range spec.Paths {
	//   for response := range op.Responses {
	//     for prop := range response.Schema.Properties where prop.format=="Money" {
	//       require prop.amount_cents.type == "integer"
	//       require prop.currency matches ISO-4217 pattern
	//     }
	//   }
	// }
}
