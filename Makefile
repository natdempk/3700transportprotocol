all:
	$(RM) 3700send
	$(RM) 3700recv
	GOPATH=${PWD}/send go build send/3700send.go
	GOPATH=${PWD}/recv go build recv/3700recv.go 

clean:
	$(RM) 3700send
	$(RM) 3700recv
