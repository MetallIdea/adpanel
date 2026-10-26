-- +goose Up
ALTER TABLE web_services ADD COLUMN working_dir TEXT NOT NULL DEFAULT '';
ALTER TABLE web_services ADD COLUMN startup_cmd TEXT NOT NULL DEFAULT '';
ALTER TABLE web_services ADD COLUMN env_content TEXT NOT NULL DEFAULT '';
ALTER TABLE web_services ADD COLUMN comment TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE web_services DROP COLUMN working_dir;
ALTER TABLE web_services DROP COLUMN startup_cmd;
ALTER TABLE web_services DROP COLUMN env_content;
ALTER TABLE web_services DROP COLUMN comment;
