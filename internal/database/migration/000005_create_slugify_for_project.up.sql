-- 1. Slugify helper (IMMUTABLE: output depends only on input)
CREATE OR REPLACE FUNCTION slugify(input text)
RETURNS text
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT trim(both '-' from
           regexp_replace(lower(input), '[^a-z0-9]+', '-', 'g'))
$$;

-- 2. Add the column as nullable so existing rows can be backfilled
ALTER TABLE "project" ADD COLUMN "slug" varchar;

-- 3. Backfill existing rows, appending -2, -3, ... to duplicates
UPDATE "project" p
SET "slug" = s.slug
FROM (
  SELECT
    x.id,
    CASE WHEN x.rn = 1 THEN x.base ELSE x.base || '-' || x.rn END AS slug
  FROM (
    SELECT
      id,
      COALESCE(NULLIF(slugify("name"), ''), 'project') AS base,
      ROW_NUMBER() OVER (
        PARTITION BY COALESCE(NULLIF(slugify("name"), ''), 'project')
        ORDER BY id
      ) AS rn
    FROM "project"
  ) x
) s
WHERE p.id = s.id;

-- 4. Enforce NOT NULL and uniqueness
ALTER TABLE "project" ALTER COLUMN "slug" SET NOT NULL;
ALTER TABLE "project"
  ADD CONSTRAINT "project_slug_unique" UNIQUE ("slug");

-- 5. Trigger function: runs on insert and on name changes
CREATE OR REPLACE FUNCTION set_project_slug()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  base      text;
  candidate text;
  n         int := 1;
BEGIN
  base := COALESCE(NULLIF(slugify(NEW."name"), ''), 'project');
  candidate := base;

  WHILE EXISTS (
    SELECT 1 FROM "project"
    WHERE "slug" = candidate
      AND "id" <> COALESCE(NEW."id", -1)
  ) LOOP
    n := n + 1;
    candidate := base || '-' || n;
  END LOOP;

  NEW."slug" := candidate;
  RETURN NEW;
END;
$$;

-- 6. Fire on insert, and on updates that touch name
CREATE TRIGGER project_slug_before_write
BEFORE INSERT OR UPDATE OF "name" ON "project"
FOR EACH ROW EXECUTE FUNCTION set_project_slug();
