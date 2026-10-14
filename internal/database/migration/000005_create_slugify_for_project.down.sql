-- Reverse order: trigger, then its function, then the column, then the helper
DROP TRIGGER IF EXISTS project_slug_before_write ON "project";
DROP FUNCTION IF EXISTS set_project_slug();

ALTER TABLE "project" DROP CONSTRAINT IF EXISTS "project_slug_unique";
ALTER TABLE "project" DROP COLUMN IF EXISTS "slug";

DROP FUNCTION IF EXISTS slugify(text);
