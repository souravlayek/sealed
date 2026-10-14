# sealed Architecture & Plan

**Project name:** sealed (CLI: `sealed`)

A secrets management system for individual developers working with collaborators on projects — similar to GitHub Secrets but for local development and CI.

> **Note:** Formerly referred to as "DropEnv" in early design docs. The product and CLI are **sealed**.

---

## 1. Scope & Users

| Aspect | Detail |
|--------|--------|
| **Users** | Individual developers, each with an account |
| **Collaboration** | A project can have multiple collaborators; **no teams/organizations in v1** |
| **Target scale** | ~50 engineers, ~10,000 CLI calls/day |
| **Availability** | High (CLI latency can be somewhat high — loads are one-time operations) |

---

## 2. Data Model

### Secret
| Field | Type | Notes |
|-------|------|-------|
| `key` | string | Environment variable name |
| `encrypted_value` | bytes | AES-256-GCM ciphertext |
| `project_id` | UUID | FK to project |
| `environment` | enum | `dev`, `staging`, `prod` |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |
| `rotation_needed_in_days` | int | Reminder only; 0 = disabled |
| `rotated_at` | timestamp | Optional explicit rotation marker (does not change value) |

### Constraints
- **Uniqueness**: `(project_id, environment, key)` is unique.
- **Key parity**: Each environment of a project *should* define the same set of keys. The dashboard warns on drift; **not enforced at write time**.
- **Rotation**: Changing a secret's value resets its rotation clock. `rotated_at` lets users record a real rotation without changing the value.

---

## 3. Encryption Design (Envelope Encryption)

### Key Hierarchy
```
KEK (Key-Encryption Key)  ── held in OpenBao transit engine, never leaves
    │
    ├─ wraps ──▶  Project DEK (256-bit random)  ── stored wrapped in project_keys table
    │                │
    │                └─ encrypts ──▶ Secret value (AES-256-GCM)
    │                     ├─ fresh 96-bit nonce per write
    │                     └─ AAD = project_id | environment | key_name
    │
    └─ KEK rotation: new key version in OpenBao → rewrap each project_keys row via rewrap endpoint
         • Secret ciphertexts **do not change**
         • AAD must **not** include KEK version
```

### DEK Rotation (rare)
- Requires `dek_version` column on `project_keys`
- Requires rewriting all secret ciphertexts for that project

### OpenBao Operational Concerns
- KEK bootstrap, unseal, and HA are operational concerns for OpenBao
- **Verify endpoint paths and response codes against the deployed OpenBao version**

---

## 4. Authentication

### CLI — OAuth Device Flow (RFC 8628)
1. CLI initiates device flow → user authorizes in browser
2. CLI generates **Ed25519 keypair**
   - Private key → stored in OS keychain
   - Public key → registered to the device
3. Each API request is **signed** over:
   - HTTP method
   - Request path
   - Timestamp
   - Nonce
   - Body hash (SHA-256)
4. Server verifies:
   - Signature validity
   - Timestamp window (~60 seconds)
   - Nonce replay (stored with TTL)
   - Device revocation status
5. **No bearer or refresh tokens** to steal

### Device Revocation
- Setting `revoked_at` on device → next request fails immediately

### Dashboard
- Separate client using normal session cookies

---

## 5. Read Path: `sealed load`

```
1. CLI → signed GET /api/v1/projects/{id}/environments/{env}/secrets
2. API verifies device signature, revocation, project membership
3. API unwraps DEK (from short-TTL in-memory cache if available)
4. API decrypts each secret with its AAD (project_id | env | key)
5. API computes `missing` list:
     keys defined in project but absent in requested environment
   (computed server-side)
6. API writes audit event:
     { key_names: [...], device_id: ..., timestamp: ... }
   **Never logs values**
7. API responds with { secrets: {...}, missing: [...] }
   • Any integrity failure → aborts whole response
   • Partial plaintext is **never** returned
```

---

## 6. CLI Behavior

### Commands
| Command | Description |
|---------|-------------|
| `sealed login` | OAuth device flow; stores Ed25519 private key in OS keychain |
| `sealed load <project> <env> [KEY...]` | Prints `export KEY='value'` lines |
| `sealed run -- <command>` | **Preferred** — injects into child process without printing values |

### `load` Output Format
- Prints `export KEY='value'` with **single quotes**
- Escapes each `'` as `'\''` (POSIX-safe)
- Values with newlines work in POSIX shells
- **NUL bytes are rejected** at write time

### Exit Codes
| Condition | Exit Code |
|-----------|-----------|
| Success, no missing | 0 |
| Missing keys, no `--allow-missing` | 3 (prints warning to stderr) |
| Missing keys, with `--allow-missing` | 0 |
| Auth / network / other error | 1 |

### Security
- CLI stores **no tokens or secrets in plaintext files**
- Private key only in OS keychain

---

## 7. Threat Model

| Threat | Impact | Mitigation |
|--------|--------|------------|
| **Database leak** | Attacker gets wrapped DEKs + ciphertexts | Both useless without KEK (in OpenBao) |
| **Stolen laptop** | Attacker has device private key | Revoke device → immediate access cut; key in OS keychain |
| **API server compromise** | Plaintext exposed while process runs | Accepted cost of server-side encryption; DEK caching is short; audit logs record activity |
| **OpenBao compromise** | KEK exposed | Mitigated by OpenBao's own unseal + access controls |
| **Removed collaborator** | Access stops immediately | Values already loaded to their machine remain → **rotate those secrets at the provider** |

---

## 8. API Endpoints (Planned)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/auth/device/start` | Start device flow |
| `POST` | `/api/v1/auth/device/poll` | Poll for device authorization |
| `GET` | `/api/v1/projects/{id}/environments/{env}/secrets` | Load secrets (signed) |
| `PUT` | `/api/v1/projects/{id}/environments/{env}/secrets/{key}` | Upsert secret (dashboard) |
| `DELETE` | `/api/v1/projects/{id}/environments/{env}/secrets/{key}` | Delete secret |
| `POST` | `/api/v1/projects/{id}/keys/rotate-dek` | Rotate DEK (admin) |
| `GET` | `/api/v1/audit` | Audit log (dashboard) |

---

## 9. Database Schema (Planned)

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Project members (collaborators)
CREATE TABLE project_members (
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member', -- 'owner', 'member'
    PRIMARY KEY (project_id, user_id)
);

-- Project DEKs (wrapped by KEK in OpenBao)
CREATE TABLE project_keys (
    project_id UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    wrapped_dek BYTEA NOT NULL,
    kek_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Secrets
CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment TEXT NOT NULL CHECK (environment IN ('dev','staging','prod')),
    key_name TEXT NOT NULL,
    encrypted_value BYTEA NOT NULL,
    nonce BYTEA NOT NULL, -- 96-bit
    rotation_needed_in_days INT NOT NULL DEFAULT 0,
    rotated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (project_id, environment, key_name)
);

-- Devices (for CLI)
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    public_key BYTEA NOT NULL, -- Ed25519
    name TEXT,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

-- Nonce store (replay protection)
CREATE TABLE used_nonces (
    nonce_hash BYTEA PRIMARY KEY, -- SHA-256 of nonce
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Audit log
CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    action TEXT NOT NULL, -- 'load', 'write', 'delete', 'rotate_dek'
    key_names TEXT[] NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
```

---

## 10. Implementation Notes for Agents

### Crypto Primitives
- **AES-256-GCM** for secret encryption (stdlib `crypto/aes` + `crypto/cipher`)
- **Ed25519** for request signatures (stdlib `crypto/ed25519`)
- **SHA-256** for body hashes and nonce hashing
- **HKDF** if key derivation needed

### OpenBao Integration
- Transit engine: `transit/encrypt/sealed-kek`, `transit/decrypt/sealed-kek`, `transit/rewrap/sealed-kek`
- Context for wrapping: `project_id` (binds DEK to project)
- Verify exact endpoint paths against deployed version

### Caching
- DEK cache: in-memory, TTL ~5 minutes, max entries ~100
- Invalidate on DEK rotation or project key update

### Audit Logging
- **Never log secret values**
- Log: `device_id`, `project_id`, `action`, `key_names[]`, `timestamp`
- Structured JSON for SIEM ingestion

### Error Handling
- Any decryption/integrity failure on `load` → **abort entire response**, return 500
- Do not return partial plaintext

---

## 11. Future Considerations (Post-v1)

- Teams / Organizations
- Secret versioning & point-in-time recovery
- Automatic rotation via provider integrations (AWS, GCP, Vercel, etc.)
- Policy engine (e.g., "prod secrets require 2-person approval")
- Secret sharing via one-time links
- CLI `push` / `pull` for syncing `.env` files

---

## 12. References

- [RFC 8628 — OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628)
- [OpenBao Transit Engine](https://openbao.org/docs/secrets/transit/)
- [NIST SP 800-38D — AES-GCM](https://csrc.nist.gov/publications/detail/sp/800-38d/final)
- [POSIX Shell Quoting](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_02_02)

---

*Document version: 1.0*  
*Project: sealed*  
*Last updated: 2026-10-14*