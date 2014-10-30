TARGETS = mysql-jsondump mysql-jsondump.Linux mysql-jsondump.Darwin
CFLAGS = -O2

default: mysql-jsondump

all: $(TARGETS)

mysql-jsondump:
	go build

mysql-jsondump.Darwin:
	go build -ccflags "$(CFLAGS)" -o $@ mysql-jsondump.go

mysql-jsondump.Linux:
	GOOS=linux GOARCH=amd64 go build -x -ccflags "$(CFLAGS)" -o $@ mysql-jsondump.go

clean:
	rm -f $(TARGETS)
