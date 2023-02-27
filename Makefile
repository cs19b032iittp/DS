tracker: sync build
	./freeflow -h 127.0.0.1 -p 4000 -r tracker -n tracker1 

user1: sync build
	./freeflow -h 127.0.0.1 -p 4001 -r user -n user1 

user2: sync build
	./freeflow -h 127.0.0.1 -p 4002 -r user -n user2

user3: sync build
	./freeflow -h 127.0.0.1 -p 4003 -r user -n user3

sync:
	go mod tidy

build:
	go build