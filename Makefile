TARGETS = sqldump sqldump.Linux sqldump.Darwin

default: sqldump

all: $(TARGETS)

sqldump: sqldump.go
	go build

sqldump.Darwin:
	go build -o $@ sqldump.go

sqldump.Linux:
	GOOS=linux GOARCH=amd64 go build -x -o $@ sqldump.go

clean:
	rm -f $(TARGETS)

get:
	go get github.com/go-sql-driver/mysql
