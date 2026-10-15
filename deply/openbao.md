# OpenBao deployment

Runs OpenBao in production mode (file storage, TLS, manual unseal). Do not use `server -dev`.

## 1. Layout

    openbao/
    ├── docker-compose.yml
    ├── config/
    │   └── openbao.hcl
    ├── tls/
    │   ├── tls.crt
    │   └── tls.key        # never committed
    └── sealed-server.hcl

## 2. Configuration

`config/openbao.hcl`:

    ui            = true

    storage "file" {
      path = "/openbao/data"
    }

    listener "tcp" {
      address       = "0.0.0.0:8200"
      tls_cert_file = "/openbao/tls/tls.crt"
      tls_key_file  = "/openbao/tls/tls.key"
    }

    api_addr     = "https://openbao:8200"
    cluster_addr = "https://openbao:8201"

Do not add `disable_mlock`; this version rejects it.

## 3. TLS certificate

Use a Tailscale-issued certificate if HTTPS certificates are enabled for your tailnet, otherwise a self-signed one. Include every name clients use in the SANs:

    openssl req -x509 -newkey rsa:4096 -nodes -days 825 \
      -keyout tls/tls.key -out tls/tls.crt \
      -subj "/CN=openbao" \
      -addext "subjectAltName=DNS:openbao,DNS:localhost,DNS:<host>.<tailnet>.ts.net,IP:127.0.0.1,IP:<tailscale-ip>"

Get the Tailscale IP with `tailscale ip -4`.

Permissions: the container runs as a non-root user and must read the key:

    chmod 644 tls/tls.crt tls/tls.key

Restrict host access by moving the key out of world-readable storage once the setup is stable.

## 4. Compose file

`docker-compose.yml`:

    services:
      openbao:
        image: quay.io/openbao/openbao:2.1.0
        entrypoint: ["bao", "server", "-config=/openbao/config/openbao.hcl"]
        cap_add:
          - IPC_LOCK
        environment:
          BAO_ADDR: https://127.0.0.1:8200
        volumes:
          - ./config:/openbao/config:ro
          - ./tls:/openbao/tls:ro
          - openbao-data:/openbao/data
        ports:
          - "<tailscale-ip>:8200:8200"
          - "127.0.0.1:8200:8200"
        networks:
          - default
          - sealed-net
        restart: unless-stopped
        healthcheck:
          test: ["CMD", "bao", "status", "-tls-skip-verify"]
          interval: 15s
          timeout: 5s
          retries: 5

    networks:
      sealed-net:
        external: true

    volumes:
      openbao-data:

Why these lines matter:

- **`entrypoint`** replaces the image's default script, which starts `server -dev` and collides with the config listener.
- **Ports bind to specific addresses.** Never publish `0.0.0.0:8200`.
- **`sealed-net`** lets the sealed server resolve `openbao`.

## 5. Start

    cd openbao
    docker compose up -d
    docker compose logs -f openbao

Expect one listener on `0.0.0.0:8200` and "security barrier not initialized". Stop following logs with Ctrl+C.

Helper alias used below (run from `openbao/`):

    E="docker compose exec -e BAO_ADDR=https://127.0.0.1:8200 -e BAO_SKIP_VERIFY=true openbao"

## 6. Initialize (once per vault)

    $E bao operator init -key-shares=5 -key-threshold=3

Save all five unseal keys and the root token immediately. This output is shown once. If the keys are lost while the vault is sealed, its data cannot be recovered.

## 7. Unseal (after every start or restart)

Run three times, with three different keys:

    $E bao operator unseal

Verify:

    curl -sk https://<tailscale-ip>:8200/v1/sys/health

You want `"sealed":false`. The sealed server cannot encrypt or decrypt while the vault is sealed.

## 8. Transit key and policy

Log in with the root token (the prompt keeps it out of shell history):

    $E bao login

Enable transit and create the KEK. `derived=true` means every call must include a `context`:

    $E bao secrets enable transit
    $E bao write -f transit/keys/sealed-kek type=aes256-gcm96 derived=true

Create the policy file on the host:

    cat > sealed-server.hcl <<'EOF'
    path "transit/encrypt/sealed-kek" { capabilities = ["update"] }
    path "transit/decrypt/sealed-kek" { capabilities = ["update"] }
    path "transit/rewrap/sealed-kek"  { capabilities = ["update"] }
    EOF

    cat sealed-server.hcl | docker compose exec -T -e BAO_ADDR=https://127.0.0.1:8200 -e BAO_SKIP_VERIFY=true openbao bao policy write sealed-server -

## 9. Scoped server token

    $E bao token create -policy=sealed-server -period=768h -orphan

Copy the `token` value into the sealed stack's `.env` as `BAO_TOKEN`.

Test it, passing a base64 context:

    docker compose exec -e BAO_ADDR=https://127.0.0.1:8200 -e BAO_SKIP_VERIFY=true -e BAO_TOKEN=<server-token> openbao \
      bao write transit/encrypt/sealed-kek plaintext=$(echo -n test | base64) context=$(echo -n 1 | base64)

A `ciphertext` in the output confirms the token, policy, and key work.

Then revoke the root token:

    $E bao token revoke <root-token>

## 10. Routine operations

| Task | Command / action |
|---|---|
| Renew server token (before 32 days) | `bao token renew` with the server token |
| Replace an expired token | Log in with root, create a new token, update `.env`, recreate `server`, revoke root |
| Unseal after restart | Three keys via `$E bao operator unseal` |
| Rotate KEK | `$E bao write -f transit/keys/sealed-kek/rotate`, then rewrap project DEKs through the server |
| Rekey unseal keys | `$E bao operator rekey -init -key-shares=5 -key-threshold=3` |
| Back up | Snapshot the `openbao-data` volume, stored apart from the keys |

## 11. Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `Error parsing listener configuration ... address already in use` | Default entrypoint starts dev listener | Use the `entrypoint` override |
| `unknown or unsupported field disable_mlock` | Field removed in this version | Delete the line |
| `permission denied` reading `tls.key` | Container user can't read key | `chmod 644 tls/tls.key` |
| `missing 'context' for key derivation` | Client omitted `context` | Send a base64 context on every transit call |
| Health returns `503` / `"sealed":true` | Restart sealed the vault | Unseal with three keys |
| Browser UI says "not available in this binary" | Official image has no UI | Use the CLI, or build a UI binary |
