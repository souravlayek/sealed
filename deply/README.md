# Sealed deployment

Sealed runs as two stacks on one host, joined by a shared Docker network:

| Stack | Guide | Contains |
|---|---|---|
| OpenBao | [openbao.md](./openbao.md) | Key management (transit KEK), unseal, scoped token |
| Sealed | [sealed.md](./sealed.md) | API server, dashboard, database connection |

**Order matters.** Set up OpenBao first, since the server needs its token and certificate.

## Prerequisites

- Docker and Docker Compose v2
- A PostgreSQL instance the sealed server can reach, with a `sealed` role and database
- A private network path for OpenBao (this guide uses Tailscale)
- A Resend account with a verified sending domain

## Shared network

Both stacks join `sealed-net`. Create it once per host:

    docker network create sealed-net

## Secrets handling

- Never commit `.env` files, unseal keys, root tokens, or the OpenBao data volume.
- Save the five unseal keys and the root token in a password manager, in at least two separate places. Do not paste them into chat, tickets, or logs.
- Rotate the root token after setup and rekey unseal keys if they were ever shared.
