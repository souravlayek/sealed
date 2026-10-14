CREATE TABLE "user" (
  "id" bigserial PRIMARY KEY,
  "email" varchar NOT NULL UNIQUE,
  "name" varchar NOT NULL,
  "role" varchar NOT NULL DEFAULT 'USER',
  "password" varchar NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "deleted_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX "user_email_active_unique"
ON "user" ("email")
WHERE "deleted_at" IS NULL;


CREATE TABLE "project" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "environments" jsonb,
  "deleted_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT now()
);


CREATE TABLE "project_members" (
  "id" bigserial PRIMARY KEY,
  "project_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "role" varchar NOT NULL,
  "deleted_at" timestamptz
);

CREATE UNIQUE INDEX "project_members_active_unique"
ON "project_members" ("project_id", "user_id")
WHERE "deleted_at" IS NULL;


CREATE TABLE "devices" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL ,
  "ed25519_pubkey" varchar NOT NULL,
  "last_used_at" timestamptz,
  "revoked_at" timestamptz,
  "deleted_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT now(),


  CONSTRAINT "devices_user_fk"
    FOREIGN KEY ("user_id")
    REFERENCES "user" ("id")
    DEFERRABLE INITIALLY IMMEDIATE
);

CREATE UNIQUE INDEX "devices_pubkey_active_unique"
ON "devices" ("ed25519_pubkey")
WHERE "deleted_at" IS NULL;


CREATE TABLE "project_keys" (
  "project_id" bigint PRIMARY KEY,
  "dek_wrapped" varchar NOT NULL,
  "kek_version" varchar NOT NULL,
  "deleted_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT "project_keys_project_fk"
    FOREIGN KEY ("project_id")
    REFERENCES "project" ("id")
    DEFERRABLE INITIALLY IMMEDIATE
);


CREATE TABLE "secrets" (
  "id" bigserial PRIMARY KEY,
  "project_id" bigint NOT NULL,
  "environment" varchar NOT NULL,
  "key_name" varchar NOT NULL,
  "ciphertext" varchar,
  "nonce" varchar,
  "kek_version" varchar,
  "value_changed_at" timestamptz,
  "rotation_days" integer,
  "created_at" timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT "secrets_project_fk"
    FOREIGN KEY ("project_id")
    REFERENCES "project" ("id")
    DEFERRABLE INITIALLY IMMEDIATE,

  CONSTRAINT "secrets_project_environment_key_unique"
    UNIQUE ("project_id", "environment", "key_name")
);


CREATE TABLE "audit_events" (
  "id" bigserial PRIMARY KEY,
  "actor_user_id" bigint,
  "device_id" bigint,
  "action" jsonb NOT NULL,
  "project_id" bigint,
  "environment" varchar NOT NULL,
  "key_names" jsonb NOT NULL,
  "ip" varchar,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz,

  CONSTRAINT "audit_events_actor_user_fk"
    FOREIGN KEY ("actor_user_id")
    REFERENCES "user" ("id")
    DEFERRABLE INITIALLY IMMEDIATE,

  CONSTRAINT "audit_events_device_fk"
    FOREIGN KEY ("device_id")
    REFERENCES "devices" ("id")
    DEFERRABLE INITIALLY IMMEDIATE,

  CONSTRAINT "audit_events_project_fk"
    FOREIGN KEY ("project_id")
    REFERENCES "project" ("id")
    DEFERRABLE INITIALLY IMMEDIATE
);


-- Useful indexes

CREATE INDEX "project_members_project_id_idx"
ON "project_members" ("project_id")
WHERE "deleted_at" IS NULL;

CREATE INDEX "project_members_user_id_idx"
ON "project_members" ("user_id")
WHERE "deleted_at" IS NULL;

CREATE INDEX "devices_user_id_idx"
ON "devices" ("user_id")
WHERE "deleted_at" IS NULL;

CREATE INDEX "secrets_project_environment_idx"
ON "secrets" ("project_id", "environment");

CREATE INDEX "audit_events_project_created_at_idx"
ON "audit_events" ("project_id", "created_at" DESC);

CREATE INDEX "audit_events_actor_created_at_idx"
ON "audit_events" ("actor_user_id", "created_at" DESC);
