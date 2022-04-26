how to run:

Installing dependencies:
1. install go 1.14 https://golang.org/doc/install
2. install docker
3. install postgres docker image

Run and build
1. run docker with postgres server
    sudo docker run --rm -p 5432:5432 --name postgres -e POSTGRES_PASSWORD=123456 postgres
2. create test db 
  psql -h localhost -U postgres
  create database cmdb;
3. build sources and install go deps
  go install ./...

linter:
installation - https://github.com/golangci/golangci-lint
run: golangci-lint run
