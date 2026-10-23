-- 0005 down: remove project slugs

DROP TRIGGER IF EXISTS project_slug_before_insert ON "project";
DROP FUNCTION IF EXISTS set_project_slug();

ALTER TABLE "project" DROP CONSTRAINT IF EXISTS "project_slug_unique";
ALTER TABLE "project" DROP COLUMN IF EXISTS "slug";

DROP FUNCTION IF EXISTS slugify(text);
