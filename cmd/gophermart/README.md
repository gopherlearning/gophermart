# cmd/gophermart

В данной директории будет содержаться код накопительной системы лояльности, который скомпилируется в бинарное
приложение.

## Запуск и отладка
```bash
METRICS=:8092 go run main.go -v 


METRICS=:8092 DATABASE_URI='postgres://appuser:appUSER1@localhost:58116/gophermart?sslmode=disable' go run main.go -v 



docker run --name gophermart-postgres \
  -e POSTGRES_DB=gophermart \
  -e POSTGRES_USER=appuser \
  -e POSTGRES_PASSWORD=appUSER1 \
  -p 0.0.0.0:58116:5432 \
  -d postgres:14

docker kill gophermart-postgres && docker rm gophermart-postgres

DROP table currencies;

docker exec -it -e PGPASSWORD=appUSER1 gophermart-postgres psql -U appuser -d gophermart






INSERT INTO users (login,hashed_password) VALUES ('test1','VERYstrong123');
INSERT INTO users (login,hashed_password) VALUES ('test2','VERYstrong123');
```