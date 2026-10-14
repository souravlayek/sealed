CREATE TABLE "device_on_hold" (
  "id" bigserial PRIMARY KEY,
  "device_code" varchar NOT NULL,
  "user_code" varchar NOT NULL,
  "device_name" varchar NOT NULL,
  "pub_key" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "expires_at" timestamptz NOT NULL
);
