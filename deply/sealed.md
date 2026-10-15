# Sealed stack deployment

Runs the API server and dashboard. Requires OpenBao and PostgreSQL to be set up first (see [openbao.md](./openbao.md)).

## 1. Prerequisites

- `sealed-net` exists and OpenBao is attached to it
- OpenBao is unsealed and has the `sealed-kek` key and a `sealed-server` token
- A PostgreSQL role and database for sealed:

      docker exec -it <postgres-container> psql -U <superuser> -c "CREATE ROLE sealed LOGIN PASSWORD '<password>';"
      docker exec -it <postgres-container> psql -U <superuser> -c "CREATE DATABASE sealed OWNER sealed;"

- The Postgres container (or host) is reachable from the sealed server. If it's on the host, `pg_hba.conf` must allow the Docker bridge subnet.

## 2. Layout

    sealed/
    ├── docker-compose.yml
    ├── .env                    # never committed
    ├── .env.example            # committed, placeholders only
    ├── sealed-ca-bundle.pem    # generated, never committed
    ├── sealed-server/
    │   ├── Dockerfile
    │   └── .dockerignore
    └── sealed-dashboard/
        ├── Dockerfile
        ├── nginx.conf
        └── .dockerignore

## 3. CA bundle

The server must trust the OpenBao certificate. Build a bundle from system CAs plus that certificate:

    cat /etc/ssl/certs/ca-certificates.crt ../openbao/tls/tls.crt > sealed-ca-bundle.pem

Regenerate it whenever the OpenBao certificate changes.

## 4. Environment

`.env.example` (commit this):

    POSTGRES_HOST=host.docker.internal
    DATABASE_URL=postgres://sealed:<password>@host.docker.internal:5432/sealed?sslmode=disable
    JWT_SECRET=<openssl rand -base64 48>
    BAO_ADDR=https://openbao:8200
    BAO_TOKEN=<scoped-token-from-openbao>
    RESEND_API_KEY=<resend-key>
    RESEND_SUPPORT_MAIL=<verified-address>
    RESEND_REPLY_TO_MAIL=<verified-address>
    FE_HOSTNAME=http://localhost:3004
    FE_SIGNUP_URL=http://localhost:3004/signup
    VITE_SERVER_URL=http://localhost:8081/api

Copy it to `.env` and fill in real values. Rules:

- `VITE_*` values are public: they are compiled into the dashboard bundle.
- `FE_HOSTNAME` and `FE_SIGNUP_URL` must match the dashboard's public URL exactly, including scheme and port.
- After changing `VITE_SERVER_URL`, rebuild the dashboard image.

## 5. Compose file

`docker-compose.yml`:

    services:
      server:
        build:
          context: ./sealed-server
        env_file:
          - .env
        environment:
          SSL_CERT_FILE: /certs/ca-bundle.pem
        volumes:
          - ./sealed-ca-bundle.pem:/certs/ca-bundle.pem:ro
        extra_hosts:
          - "host.docker.internal:host-gateway"
        ports:
          - "8081:8080"
        networks:
          - sealed-net
        restart: unless-stopped

      dashboard:
        build:
          context: ./sealed-dashboard
          args:
            VITE_SERVER_URL: ${VITE_SERVER_URL}
        ports:
          - "3004:80"
        depends_on:
          - server
        networks:
          - sealed-net
        restart: unless-stopped

    networks:
      sealed-net:
        external: true

Use `extra_hosts` only when the database is on the host. Remove it if the database is a container on `sealed-net`.

## 6. Start

    cd sealed
    docker compose up -d --build
    docker compose logs -f server

Apply schema migrations to the `sealed` database before the first sign-in.

## 7. Verify

Server reaches OpenBao:

    docker compose exec server wget -qO- https://openbao:8200/v1/sys/health; echo

Want `"sealed":false`. A `"sealed":true` means unseal OpenBao first.

API responds:

    curl -i http://localhost:8081/api/v1/device/code -H 'Content-Type: application/json' -d '{}'

Want a `400` validation response, not a connection error.

Dashboard: open `http://localhost:3004`, sign up, and create a project.

## 8. Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `connection refused` to `localhost:5432` | `localhost` is the container | Use `host.docker.internal` or the container name |
| `password authentication failed` | Role password mismatch | Match `.env` to the role |
| `no such host openbao` | Server not on `sealed-net` | Add `sealed-net` to the service and OpenBao |
| `x509: certificate signed by unknown authority` | CA bundle missing or `SSL_CERT_FILE` unset | Check the mount and env var |
| `x509: ... not openbao` | Certificate SANs lack `openbao` | Regenerate the certificate with `DNS:openbao` |
| `403` from OpenBao | Stale or wrong `BAO_TOKEN` | Update `.env`, recreate `server` |
| `missing 'context' for key derivation` | Transit call without context | Send base64 context on encrypt, decrypt, rewrap |
| Dashboard calls wrong API | `VITE_SERVER_URL` changed without rebuild | `docker compose up -d --build dashboard` |
| CORS errors | `FE_HOSTNAME` doesn't match dashboard URL | Align `FE_HOSTNAME` and the port |

## 9. Operations

- **Restart checklist:** after any host or OpenBao restart, unseal OpenBao, then confirm `sealed:false` before using the app.
- **Token renewal:** renew the `sealed-server` token before its 32-day period ends. See [openbao.md](./openbao.md), section 10.
- **Backups:** back up the `sealed` database and the `openbao-data` volume separately.
- **Secret rotation:** rotate the Resend key and `JWT_SECRET` if either is exposed. Rotating `JWT_SECRET` logs everyone out.
