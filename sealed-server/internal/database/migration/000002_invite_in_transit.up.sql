CREATE TABLE "project_invites" (
  "id" bigserial PRIMARY KEY,
  "invited_email" varchar UNIQUE NOT NULL,
  "project_id" bigint NOT NULL
);

ALTER TABLE "project_invites" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("id") DEFERRABLE INITIALLY IMMEDIATE;
