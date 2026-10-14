ALTER TABLE "devices"
  ADD COLUMN "device_name" varchar,
  ADD COLUMN "device_code" varchar;

ALTER TABLE "devices"
  ADD CONSTRAINT "devices_device_code_unique" UNIQUE ("device_code");
