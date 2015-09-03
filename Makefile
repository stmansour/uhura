all: uhura
	echo "*** COMPLETED ***"

install: uhura
	cp uhura /usr/local/accord/bin

uhura: *.go
	go build


