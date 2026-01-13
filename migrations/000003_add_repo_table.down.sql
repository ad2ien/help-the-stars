DROP TABLE repos;

ALTER TABLE issues ADD COLUMN repo_with_owner text;
ALTER TABLE issues ADD COLUMN repo_description text;
ALTER TABLE issues ADD COLUMN stargazers_count text;
