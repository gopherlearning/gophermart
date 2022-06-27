--- +
CREATE TABLE users (  
  id BIGSERIAL UNIQUE PRIMARY KEY,
  login VARCHAR ( 50 ) NOT NULL UNIQUE CHECK(length(login) >= 4),
  hashed_password VARCHAR ( 250 ) NOT NULL CHECK(length(hashed_password) >= 20)
);
--- +
CREATE TABLE sessions (  
  id BIGSERIAL UNIQUE PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  claim JSONB NOT NULL
);
--- +
CREATE TABLE orders (  
  id BIGINT UNIQUE PRIMARY KEY,
  accrual REAL,
  created_at TIMESTAMPTZ NOT NULL,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
--- +
CREATE TABLE order_statuses (  
  status INT  NOT NULL,
  order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (order_id, status)
);

CREATE TABLE withdraws (  
  id BIGINT UNIQUE PRIMARY KEY,
  sum REAL NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
