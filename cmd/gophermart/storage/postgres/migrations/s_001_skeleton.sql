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
  id BIGSERIAL UNIQUE PRIMARY KEY,
  sum REAL NOT NULL,
  order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL
);



SELECT o.id, o.accrual, o.created_at, s.status
FROM orders AS o 
JOIN order_statuses AS s 
  ON o.id = s.order_id 
  AND s.created_at = (select max(created_at) from order_statuses where order_id=o.id);

select o.id, o.accrual, o.created_at, s.status
from documents
left join updates
  on updates.document_id=documents.id
  and updates.date=(select max(date) from updates where document_id=documents.id)
where documents.id=?;