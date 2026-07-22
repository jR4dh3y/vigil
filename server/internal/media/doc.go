// Package media owns the media plane: MediaMTX path config, short-lived stream
// tokens, the MediaMTX external-auth hook, and FFmpeg utilities (snapshots).
//
// Clients never hold standing MediaMTX credentials. POST /cameras/{id}/live
// mints a ~60s reusable token; MediaMTX calls POST /internal/mediamtx/auth to
// validate it. See arch.md §3 and §7.
package media
