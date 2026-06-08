// Package eidas is the Regime A (signing & evidence) audit emitter for the
// eSignature Portal. It gives every signing-event service one correct way to
// publish structured, schema-correct material-signing events to the broker, so
// the hash-chained trail consumed by the Audit & Evidence Service is consistent
// regardless of which service produced the event (Audit Design §3, §8.1;
// Services Catalog §3.9.4).
//
// It is a thin layer over go-platform-kit's broker: every event is the frozen
// §8.1 broker.Envelope, tagged broker.CategorySigning, stamped with a ULID event
// id, a high-precision occurrence time, and the request's correlation/trace ids
// (broker.Publish handles stamping + validation), then published to the signing
// topic over the broker's TLS + per-topic ACLs. Async is fine here — integrity,
// not latency, is the priority (Audit Design §8).
//
// # Lean / reference-only (Audit Decisions D2)
//
// The portal is not a certified Qualified validation service. The cryptographic
// evidence lives with LVRTC and inside the self-contained B-LTA file; this
// stream stores only the minimum needed to find and trust that evidence —
// material-event records plus references (envelope/slot, session id, signature
// type, at-signing validation pass/fail, qualified-timestamp presence, the S3
// ref + content hash of the signed file). It must never carry certificates,
// digests, OCSP/CRL, archive material, full validation blobs, or document bytes;
// the emitter strips such "fat" attribute keys defensively (see sanitize).
//
// Regime B (GDPR personal-data access) and Regime C (NIS2 security telemetry) are
// separate mechanisms with their own libraries (go-gdpr-audit, go-sec-events) —
// a service that uploads a document emits this Regime A event *and* the
// corresponding access record, rather than overloading one event (Audit Design
// §8, §11).
package eidas
