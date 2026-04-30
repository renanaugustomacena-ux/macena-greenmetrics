# Packs

This directory holds Packs — the swappable per-engagement, per-flavour
implementations of Core's Pack-contract interfaces. The Pack model is
described in `docs/MODULAR-TEMPLATE-CHARTER.md` §3 (Core vs. Pack), §4 (Pack
Contract), and `docs/DOCTRINE.md` Rules 69–88 (Modular Template Integrity).

## Layout

```
packs/
├── README.md                    (this file)
├── protocol/                    Protocol Packs — implement EP-01 Ingestor
│   ├── modbus_tcp/
│   ├── modbus_rtu/
│   ├── mbus/
│   ├── sunspec/
│   ├── pulse/
│   ├── ocpp_1_6/
│   ├── ocpp_2_0_1/
│   ├── iec_61850/
│   ├── opc_ua/
│   ├── mqtt_sparkplug_b/
│   └── bacnet/
├── factor/                      Factor Packs — implement EP-03 FactorSource
│   ├── ispra/
│   ├── gse/
│   ├── terna/
│   ├── aib/
│   ├── uk_defra/
│   └── epa_egrid/
├── report/                      Report Packs — implement EP-02 Builder
│   ├── esrs_e1/
│   ├── piano_5_0/
│   ├── conto_termico/
│   ├── tee/
│   ├── audit_dlgs102/
│   ├── monthly_consumption/
│   ├── co2_footprint/
│   ├── uk_secr/
│   ├── ghg_protocol/
│   ├── iso_14064_1/
│   ├── tcfd/
│   └── ifrs_s_1_s_2/
├── identity/                    Identity Packs — replace local-DB auth
│   ├── saml/
│   └── oidc/
└── region/                      Region Packs — bundle regional defaults
    ├── it/
    ├── de/
    ├── fr/
    ├── es/
    ├── gb/
    └── at/
```

This layout is the *target* for the end of Phase H Sprint S17. The current
state populates only the directories already extracted from `internal/services/`.
The extraction work is the Phase E Sprint S6–S7 deliverable per `docs/PLAN.md`.

## Pack contents (per-Pack)

Each Pack directory carries:

```
packs/<kind>/<id>/
├── manifest.yaml                Per Rule 70; schema at docs/contracts/pack-manifest.schema.json
├── CHARTER.md                   Per Rule 87 acceptance; carries the Tradeoff Stanza
├── README.md                    Human-readable Pack overview
├── pack.go                      Implements internal/packs.Pack
├── <kind>_specific.go           Per-kind contract implementation (Ingestor / Builder / ...)
├── config/
│   └── defaults.yaml            Pack-scoped config defaults (Rule 82)
├── tests/
│   ├── unit_test.go             Pack-specific unit tests
│   ├── conformance_test.go      Conformance against the Pack-kind contract (Rule 84)
│   └── fixtures/                Synthetic fixtures only (Rule 165)
├── simulator/                   For Protocol Packs only (Rule 121)
│   └── main.go
└── (kind-specific)/
    ├── devices/                 For Protocol Packs (Rule 122)
    ├── xsd/                     For Report Packs requiring XSD validation
    ├── taxonomy.yaml            For Report Packs requiring taxonomy mapping
    └── thresholds.yaml          For Region Packs (Rule 139)
```

## Pack acceptance

A Pack is accepted into upstream only if it passes the checklist in
`docs/PACK-ACCEPTANCE.md` (Phase E Sprint S5 deliverable):

1. `manifest.yaml` validates against `docs/contracts/pack-manifest.schema.json`.
2. `CHARTER.md` carries the four-part Tradeoff Stanza.
3. The Pack-kind conformance test passes against the Core-supported window of
   contract versions.
4. The Pack imports no other Pack outside its declared `dependencies` field.
5. At least one Macena platform-team reviewer has built it from clean clone in
   under 30 minutes.

## Pack lifecycle in the loader

See `backend/internal/packs/pack.go` for the lifecycle contract. Briefly:

1. Loader reads the manifest.
2. Loader validates against schema.
3. Loader checks `min_core_version` and `pack_contract_version`.
4. Loader instantiates via the Pack's `New()` constructor.
5. Loader calls `Pack.Init(ctx, core)`.
6. Loader calls `Pack.Register(reg)`.
7. Loader records the Pack in `manifest.lock.json` (Cosign-signed).
8. At runtime, `/api/health` invokes `Pack.Health(ctx)`.
9. On graceful shutdown, `Pack.Shutdown(ctx)` runs with a 30-second budget.
