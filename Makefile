APP     := smtptest
DIST    := dist
ZIP     := $(APP).zip

.PHONY: all linux windows zip clean

all: linux windows

linux:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(APP) .

windows:
	mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(APP).exe .

zip: linux windows
	rm -f $(ZIP)
	zip -j $(ZIP) $(DIST)/$(APP) $(DIST)/$(APP).exe

clean:
	rm -rf $(DIST) $(ZIP)
