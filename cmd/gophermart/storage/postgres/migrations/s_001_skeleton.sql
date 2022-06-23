--- +
CREATE TABLE users (  
  id BIGSERIAL UNIQUE PRIMARY KEY,
  login VARCHAR ( 50 ) NOT NULL UNIQUE,
  hashed_password VARCHAR ( 250 ) NOT NULL
);
--- +
CREATE TABLE sessions (  
  id BIGSERIAL UNIQUE PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  claim JSONB NOT NULL
);
--- +
CREATE TABLE orders (  
  id INT UNIQUE PRIMARY KEY,
  accrual REAL,
  created_at TIMESTAMPTZ NOT NULL,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
--- +
CREATE TABLE order_statuses (  
  id BIGSERIAL UNIQUE PRIMARY KEY,
  status INT  NOT NULL,
  order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE withdraws (  
  id BIGSERIAL UNIQUE PRIMARY KEY,
  sum REAL NOT NULL,
  order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL
);
