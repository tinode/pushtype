# pushtype — Common push types and error definitions 🔧

## TL;DR
This package contains shared type definitions and constants used by both the Tinode Push Gateway (TNPG) and the FCM push implementation. It centralizes payload shapes (notifications/data), configuration helpers, and canonical error codes returned by push adapters.

## Contents
- **Payload** — notification payload fields used for Android, iOS (APNS), and web pushes.
- **Config** — per-notification-type defaults and helpers (`GetStringField`, `GetIntField`).
- **TNPGResponse** — standardized per-message response returned by the push gateway adapters (fields: `MessageID`, `Code`, `ErrorCode`, `ErrorMessage`, `Index`).
- **Action constants** — `ActMsg`, `ActSub`, `ActRead` (message, subscription, read).
- **Platform enums/constants** — Android visibility, notification priorities, APNS headers and push types.
- **FCM error constants** — canonical string error codes (e.g., `ErrorUnregistered`, `ErrorInvalidArgument`, `ErrorQuotaExceeded`) used by adapters to normalize Google API errors.

## Usage
Import the package and reference the shared types:

```go
import "github.com/tinode/pushtype"

var p pushtype.Payload

// or a TNPG response
var resp pushtype.TNPGResponse
```

Adapters should use the constants here to map provider errors to stable values consumed by the rest of the system.

## Notes & Best Practices 💡
- Keep mapping logic in adapters (e.g., decoding Google API errors) and map to the constants in this package so the rest of the system has a stable error vocabulary.
- Use `Config` helpers to centralize defaults and avoid duplicated field lookups across adapters.
- When adding new error codes, include a short comment describing when it is used and how to recover (if applicable).

If you want, I can add a short example showing how to map a Google API error into a `TNPGResponse` using the constants in this package. ✅
