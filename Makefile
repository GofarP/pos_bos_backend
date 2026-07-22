include .env
export

migrate-up:
	$(shell go env GOPATH)/bin/migrate -path database/migrations -database "mysql://root:@tcp(127.0.0.1:3306)/pos_bos" -verbose up

migrate-down:
	$(shell go env GOPATH)/bin/migrate -path database/migrations -database "mysql://root:@tcp(127.0.0.1:3306)/pos_bos" -verbose down

migrate-force:
	@read -p "Enter version: " version; \
	$(shell go env GOPATH)/bin/migrate -path database/migrations -database "mysql://root:@tcp(127.0.0.1:3306)/pos_bos" force $$version
