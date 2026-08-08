# Decision — API gateway with shared Bearer secret (2026-08-08)

## Problem
The backend API is directly reachable at the server origin. The frontend (hosted on Vercel) is the only intended client, but today any script can call the API without a key.

## Decision
Add a gateway at the HTTP layer that requires `Authorization: Bearer <SECRET>` on every `/api/*` request (excluding `/api/health` for probes). The frontend never embeds the secret: a Vercel serverless proxy / local Vite dev proxy injects the header. This keeps the secret off the public bundle while blocking direct calls to the backend origin that lack it.

## Details
- Middleware in `src/api/gate_middleware.go`, wrapping the whole mux in `main.go`, applying only to `/api/` prefixed paths (health exempt; `/media`, `/` exempt).
- Constant-time comparison (`crypto/subtle.ConstantTimeCompare`) of the token.
- If `API_GATEWAY_SECRET` is not set: gateway is a no-op + startup warning (dev-friendly; production must set it via server/docker env and the deploy workflow secret).

## Scope / non-goals
- Does not replace session auth; it adds the first layer above it.
- `/media/*` stays public (announcement images are public by design).

## Limitations
- Ready Server remains HTTP on port 8090 with no Nginx/HTTPS. Bearer credentials can be intercepted in transit; a future HTTPS deployment must rotate the shared secret.
