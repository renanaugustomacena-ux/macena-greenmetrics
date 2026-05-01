// White-label branding facade.
//
// Re-exports the build-time-generated branding tokens from
// `branding.generated.ts`. Components and pages should import from this
// file (NOT directly from the generated file); future white-label
// extensions (runtime overrides via cookies, A/B brand tests, etc.) hook
// in here without breaking call-sites.
//
// Doctrine refs: Rule 80 (white-label conformance), Rule 82 (engagement
// overlay supersedes Core defaults).
// Charter ref: §6 (white-label readiness).
//
// EDIT: change values in `config/branding.yaml`, then run
//       `npm run gen:branding` (or just `npm run build` / `npm run dev`).

export {
	branding,
	productName,
	productTagline,
	productDescription,
	vendorName,
	supportEmail,
	primaryColor,
	accentColor,
	backgroundColor,
	logoPath,
	faviconPath,
	footerHTML,
	showPoweredBy,
} from './branding.generated';

export type { Branding } from './branding.generated';
