// Package v1 holds request/response DTOs and the centralised validator + bind helpers.
//
// Doctrine refs: Rule 14 (contract first), Rule 34 (backend contracts), Rule 39 (security as core).
// ADR: docs/adr/0012-validator-go-playground.md, docs/adr/0013-oapi-codegen-design-first.md.
package v1

import (
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// validate is a process-wide validator instance. Custom validators registered once.
var (
	validate     *validator.Validate
	validateOnce sync.Once
)

// Validator returns the shared validator instance, initialised lazily on first call.
//
// Custom validators registered:
//   - iso4217: ISO-4217 currency code (3 uppercase letters)
//   - rfc3339utc: RFC 3339 timestamp with explicit UTC offset (Z or +00:00)
//   - uuidv4: UUIDv4 string
//   - tenant_id: alias for uuidv4 with semantic meaning
//   - pod_code: Italian POD electricity supply code (regex IT001E[0-9A-Z]{8})
//   - pdr_code: Italian PDR gas supply code (14 digits)
//   - cf: Italian codice fiscale (16 chars uppercase)
//   - piva: Italian partita IVA (11 digits)
func Validator() *validator.Validate {
	validateOnce.Do(func() {
		v := validator.New(validator.WithRequiredStructEnabled())
		mustRegister(v, "iso4217", validateISO4217)
		mustRegister(v, "rfc3339utc", validateRFC3339UTC)
		mustRegister(v, "uuidv4", validateUUIDv4)
		mustRegister(v, "tenant_id", validateUUIDv4)
		mustRegister(v, "pod_code", validatePODCode)
		mustRegister(v, "pdr_code", validatePDRCode)
		mustRegister(v, "cf", validateCF)
		mustRegister(v, "piva", validatePIVA)
		validate = v
	})
	return validate
}

func mustRegister(v *validator.Validate, tag string, fn validator.Func) {
	if err := v.RegisterValidation(tag, fn); err != nil {
		// Registration failure at init is unrecoverable; panic at process start, never on user input.
		panic("validator: failed to register " + tag + ": " + err.Error())
	}
}

// ----- custom validators -----------------------------------------------------

var iso4217Set = map[string]struct{}{
	"EUR": {}, "USD": {}, "GBP": {}, "CHF": {}, "JPY": {}, "CAD": {}, "AUD": {},
	"NZD": {}, "SEK": {}, "NOK": {}, "DKK": {}, "PLN": {}, "CZK": {}, "HUF": {},
	"RON": {}, "BGN": {}, "HRK": {}, "RUB": {}, "TRY": {}, "CNY": {}, "HKD": {},
	"SGD": {}, "INR": {}, "BRL": {}, "MXN": {}, "ZAR": {}, "AED": {}, "SAR": {},
	// Extend as needed; full ISO-4217 list in vendor data file (S5).
}

func validateISO4217(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) != 3 {
		return false
	}
	upper := strings.ToUpper(s)
	if upper != s {
		return false
	}
	_, ok := iso4217Set[upper]
	return ok
}

func validateRFC3339UTC(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return false
	}
	// Require UTC offset explicitly — Z or +00:00 / -00:00.
	_, offset := t.Zone()
	return offset == 0
}

func validateUUIDv4(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	id, err := uuid.Parse(s)
	if err != nil {
		return false
	}
	return id.Version() == 4
}

var podCodeRegex = regexp.MustCompile(`^IT001E[0-9A-Z]{8}$`)

func validatePODCode(fl validator.FieldLevel) bool {
	return podCodeRegex.MatchString(fl.Field().String())
}

var pdrCodeRegex = regexp.MustCompile(`^[0-9]{14}$`)

func validatePDRCode(fl validator.FieldLevel) bool {
	return pdrCodeRegex.MatchString(fl.Field().String())
}

var cfRegex = regexp.MustCompile(`^[A-Z0-9]{16}$`)

func validateCF(fl validator.FieldLevel) bool {
	return cfRegex.MatchString(strings.ToUpper(fl.Field().String()))
}

var pivaRegex = regexp.MustCompile(`^[0-9]{11}$`)

func validatePIVA(fl validator.FieldLevel) bool {
	return pivaRegex.MatchString(fl.Field().String())
}

// ValidationErrors maps validator.ValidationErrors into a flat slice for RFC 7807
// detail field assembly. Each entry: "field <Field>: failed <Tag>(<Param>)".
func ValidationErrors(err error) []string {
	if err == nil {
		return nil
	}
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []string{err.Error()}
	}
	out := make([]string, 0, len(ve))
	for _, e := range ve {
		if e.Param() != "" {
			out = append(out, "field "+e.StructField()+": failed "+e.Tag()+"("+e.Param()+")")
		} else {
			out = append(out, "field "+e.StructField()+": failed "+e.Tag())
		}
	}
	return out
}
