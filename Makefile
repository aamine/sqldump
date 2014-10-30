TARGETS = sqldump sqldump.Linux sqldump.Darwin
CFLAGS = -O2

default: sqldump

all: $(TARGETS)

sqldump: sqldump.go
	go build

sqldump.Darwin:
	go build -ccflags "$(CFLAGS)" -o $@ sqldump.go

sqldump.Linux:
	GOOS=linux GOARCH=amd64 go build -x -ccflags "$(CFLAGS)" -o $@ sqldump.go

clean:
	rm -f $(TARGETS)
