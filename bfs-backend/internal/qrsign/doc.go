// Package qrsign produces and verifies self-contained, Ed25519-signed QR tokens
// a pickup station can verify offline.
//
// This is the wire-format contract; the web (@noble/ed25519) and Android
// verifiers MUST mirror it byte for byte.
//
//	alg     Ed25519 (RFC 8032): 32-byte public key, 64-byte signature
//	body    canonical JSON, FIXED key order (marshal the Payload struct, never a map):
//	          {"v":1,"o":"<orderId>","i":<unixSeconds>,"e":<unixSeconds>,"l":[{"p":"<productId>","q":<qty>}]}
//	        i = issued-at, e = expires-at (end of the issue day, Europe/Zurich);
//	        l = redeemable physical items: simple products + chosen menu components
//	token   base64url-nopad( body || sig ) — sig is the LAST 64 bytes
//
// Verify: base64url-decode, split off the trailing 64-byte sig, ed25519.Verify
// over the LITERAL body bytes (never re-serialize), unmarshal, reject any unknown
// `v`, then reject once the verifier's clock is past `e` (fail closed if `e` is
// missing). Single-tenant: no tenant id in the payload.
//
// A token is signed only when its order is paid, and is valid only through the
// end of the day it was issued — the next day it no longer redeems.
//
// The public key is served at GET /v1/qr-config as base64url-nopad of the 32 raw
// bytes; the seed comes from QR_ED25519_PRIVATE_SEED and is never exposed.
package qrsign
