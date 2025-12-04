start mysql redis rabbitmq
service mysql start
service redis-server start
service rabbitmq-server start

go run ./cmd/admin
go run ./cmd/web