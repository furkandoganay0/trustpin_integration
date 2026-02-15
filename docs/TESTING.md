# Testing Guide

## Quick start (local, in-memory)

1) Create env file from example and fill JWT keys:

   - Copy `config/.env.example` to `config/.env`
   - Put your PEM keys into `JWT_PUBLIC_KEY` and `JWT_PRIVATE_KEY` (single line, with `\n`)

2) Load env vars and run the server:

   ```bash
   cd /Users/furkandoganay/Desktop/projects/codex/trustpin_integration
   set -a
   source /Users/furkandoganay/Desktop/projects/codex/trustpin_integration/config/.env
   set +a

   go run /Users/furkandoganay/Desktop/projects/codex/trustpin_integration/cmd/server/main.go
   ```

3) Test endpoints:

   - Swagger UI: http://localhost:8080/swagger/
   - Health: http://localhost:8080/healthz

   Import Postman collection:
   - `docs/postman_collection.json`

## MFA flow order

1) Login -> `access_token`
2) Enroll -> `pairing_code`
3) Activate
4) Create Challenge -> `challenge_id`
5) Approve
6) Get Challenge
7) Get Status

## Required headers for MFA endpoints

All `/api/mfa/*` endpoints require **both** headers:

- `Authorization: Bearer <token>`
- `X-Tenant-ID: <tenant_id>` (must match the token's `tenant_id` claim)

If `X-Tenant-ID` is missing or mismatched, the API returns **403**.

## Notes

- If `DB_DSN` and `REDIS_ADDR` are empty, the app runs in-memory (demo user is seeded).
- MFA endpoints call TrustPin. Set `TRUSTPIN_API_KEY` for successful MFA flows.
