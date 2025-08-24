CREATE TABLE
  issues (
    url text PRIMARY KEY,
    repo_with_owner text NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    creation_date timestamp NOT NULL,
    repo_description text NOT NULL,
    stargazers_count integer
  );