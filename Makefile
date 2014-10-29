CFLAGS = -O2

all: mysql-jsondump mysql-jsondump.Linux mysql-jsondump.Darwin

mysql-jsondump:
	go build

mysql-jsondump.Darwin:
	go build -ccflags "$(CFLAGS)" -o $@ mysql-jsondump.go

mysql-jsondump.Linux:
	GOOS=linux GOARCH=amd64 go build -x -ccflags "$(CFLAGS)" -o $@ mysql-jsondump.go
