CREATE TABLE
  issues (
    id INTEGER PRIMARY KEY,
    link text NOT NULL,
    title text,
    description text,
    owner text,
    creation_date timestamp
  );