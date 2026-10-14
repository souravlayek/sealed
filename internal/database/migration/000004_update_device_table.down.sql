ALTER TABLE "devices"
  DROP CONSTRAINT IF EXISTS "devices_device_code_unique";

ALTER TABLE "devices"
  DROP COLUMN IF EXISTS "device_code",
  DROP COLUMN IF EXISTS "device_name";
