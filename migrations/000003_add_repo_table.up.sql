CREATE TABLE
  repos (
    repo_with_owner text PRIMARY KEY,
    description text,
    stargazers_count integer,
    language text
  );

-- remove unused columns from issues
ALTER TABLE issues DROP COLUMN repo_description;
ALTER TABLE issues DROP COLUMN stargazers_count;

DELETE FROM issues;

-- Create foreign key issues_repos_FK
CREATE TEMPORARY TABLE temp AS
SELECT
  url,
  repo_with_owner,
  title,
  description,
  creation_date
FROM issues;

DROP TABLE issues;

CREATE TABLE issues (
	url TEXT PRIMARY KEY,
	repo_with_owner TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	creation_date TIMESTAMP NOT NULL,
	CONSTRAINT issues_repos_FK FOREIGN KEY (repo_with_owner) REFERENCES repos(repo_with_owner)
);

INSERT INTO issues
 (url,
  repo_with_owner,
  title,
  description,
  creation_date)
SELECT
  url,
  repo_with_owner,
  title,
  description,
  creation_date
FROM temp;

DROP TABLE temp;
