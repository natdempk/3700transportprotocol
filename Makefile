.ONESHELL:
all:
	$(RM) 3700send
	$(RM) 3700recv
	cd send3700
	export GOPATH=${PWD}/send3700
	go build
	cd ../recv3700
	export GOPATH=${PWD}/recv3700
	go build
	mv send3700/3700send .
	mv recv3700/3700recv .

clean:
	$(RM) 3700send
	$(RM) 3700recv
