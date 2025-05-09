CREATE TABLE accounts (
  id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  username TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL,
  picture TEXT,
  bio TEXT,
  created TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE sets (
  id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  account_id INT REFERENCES accounts(id) ON DELETE CASCADE NOT NULL,
  name TEXT,
  description TEXT,
  created TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE cards (
  id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  set_id INT REFERENCES sets(id) ON DELETE CASCADE NOT NULL,
  front TEXT,
  back TEXT,
  created TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE refreshtokens (
  id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  account_id INT REFERENCES accounts(id),
  token TEXT NOT NULL UNIQUE,
  expires TIMESTAMPTZ NOT NULL
);
 