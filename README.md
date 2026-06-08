# go-eidas-audit

The **Regime A** (signing & evidence) audit emitter for the eSignature Portal. A thin
library that gives every signing-event service **one correct way** to publish
structured, schema-correct material-signing events to the broker, so the hash-chained
trail consumed by the **Audit & Evidence Service** is consistent regardless of producer.

See the [Audit Design](../eSignature-Portal-Audit-Design.md) (§3 Regime A, §8.1 envelope)
and [Services Catalog](../eSignature-Portal-Services-Catalog.md) §3.9.4 for the design.

It builds on [`go-platform-kit`](../go-platform-kit)'s `broker`: every event is the
**frozen §8.1 `broker.Envelope`**, tagged `signing`, stamped with a ULID id, a
high-precision occurrence time, and the request's correlation/trace ids, then published to
the signing topic. Async is fine here — **integrity, not latency**, is the priority.

> **Lean / reference-only (Audit Decisions D2).** The portal is *not* a certified Qualified
> validation service. The cryptographic evidence lives with LVRTC and inside the
> self-contained B-LTA file; this stream stores only the **minimum to find and trust** that
> evidence — material events + references. It must **never** carry certificates, digests,
> OCSP/CRL, archive material, full validation blobs, or document bytes. The emitter strips
> such "fat" attribute keys defensively, and the publisher strips bearer-token-shaped keys.

## Install

```sh
go get github.com/gmb-sig/go-eidas-audit
```

Pinned in lockstep to `github.com/gmb-sig/go-platform-kit` (which pins `azugo.io/*` v0.32.x).

## Usage

Construct one `Emitter` per service over the service's `broker.Publisher` (the publisher
carries the injected transport — TLS + per-topic ACLs — and stamps every event):

```go
import (
    "github.com/gmb-sig/go-eidas-audit/eidas"
    "github.com/gmb-sig/go-platform-kit/broker"
)

pub := broker.NewPublisher(transport, cfg.ServiceName) // transport wired by the service
audit := eidas.NewEmitter(pub, eidas.DefaultTopic)      // or a configured topic
```

Then emit material signing events with the typed helpers — the event type, category,
operation and lean attribute shape are fixed for you:

```go
func (r *router) upload(ctx *azugo.Context) {
    // …store the document, compute its hash…
    err := audit.DocumentUploaded(ctx, eidas.DocumentUpload{
        Actor:        broker.Actor{ID: ctx.User().ID(), Type: "user"},
        DataSubjects: []string{ctx.User().ID()},
        DocumentID:   doc.ID,
        ContentHash:  doc.SHA256,
        MIME:         doc.MIME,
        Size:         doc.Size,
    })
    // …
}

audit.SignatureApplied(ctx, eidas.Signature{
    Actor:                 broker.Actor{ID: signer, Type: "user", Assurance: "high"},
    EnvelopeID:            env.ID,
    Slot:                  slot.ID,
    Format:                eidas.FormatPAdES,
    Level:                 eidas.LevelQES,
    SimpleSignRef:         resp.Reference, // a reference, not crypto material
    BLTAConfirmed:         true,
    QualifiedTimestampRef: resp.TSARef,
})
```

For events not covered by a helper, use the `Emit` escape hatch — it defaults the category
to `signing` and runs the same fat/PII + token sanitization before publishing.

## Events

One typed helper per Regime A material event (Audit Design §3, §7):

| Helper | `event_type` |
|---|---|
| `DocumentUploaded` | `document.uploaded` |
| `DocumentPreviewed` | `document.previewed` |
| `ConsentCaptured` | `signing.consent` |
| `AuthAssurance` | `signing.assurance` |
| `SigningInitiated` | `signing.initiated` |
| `ProviderRedirect` / `ProviderCallback` | `signing.redirect` / `signing.callback` |
| `SignatureApplied` | `signing.applied` |
| `ValidationPerformed` | `signing.validation` |
| `EnvelopeTransition` | `envelope.transition` |
| `CoSignerInvited` | `envelope.cosigner_invited` |
| `DocumentDownloaded` | `document.downloaded` |
| `RetentionPurge` | `retention.purge` |

Regime B (GDPR access) and Regime C (NIS2 security) are **separate mechanisms** with their
own libraries — [`go-gdpr-audit`](../go-gdpr-audit) and [`go-sec-events`](../go-sec-events).
A service that uploads a document emits the Regime A event here *and* the corresponding
access record there; this library does not route across regimes.

## Develop

```sh
go build ./...
go test ./...
go vet ./...
```

## License

MIT — see [LICENSE](./LICENSE).
