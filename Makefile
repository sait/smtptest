APP     := smtptest
DIST    := dist

.PHONY: all linux windows clean

all: linux windows

linux:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(APP) .

windows:
	mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(APP).exe .

clean:
	rm -rf $(DIST)
