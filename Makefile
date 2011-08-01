GOPATH=$(shell pwd)
GOINSTALL_FLAGS=-dashboard=false -clean=true

VERSION=0.2.0-1

all: build

build:
	GOPATH=$(GOPATH) goinstall $(GOINSTALL_FLAGS) conductor
	GOPATH=$(GOPATH) goinstall $(GOINSTALL_FLAGS) player
	GOPATH=$(GOPATH) goinstall $(GOINSTALL_FLAGS) submitjob
	GOPATH=$(GOPATH) goinstall $(GOINSTALL_FLAGS) getstatus

deb:	build-root
	fpm -s dir -t deb \
		-n 'orchestra' \
		-v "$(VERSION)" \
		-m 'Chris Collins <chris.collins@anchor.net.au>' \
		-p "../orchestra-$(VERSION).deb" \
		--description "Services for getting stuff run" \
		--url "http://github.com/kuroneko/orchestra/" \
		--conflicts 'orchestra-conductor, orchestra-player-go' \
		--replaces 'orchestra-conducutor, orchestra-player-go' \
		--config-files etc/conductor/players \
		-C build-root \
		.

rpm:	build-root
	fpm -s dir -t rpm \
		-n 'orchestra' \
		-v "$(VERSION)" \
		-m 'Chris Collins <chris.collins@anchor.net.au>' \
		-p "../orchestra-$(VERSION).rpm" \
		--description "Services for getting stuff run" \
		--url "http://github.com/kuroneko/orchestra/" \
		--config-files etc/conductor/players \
		-C build-root \
		.

build-root:	build
	rm -rf build-root
	mkdir -p build-root/usr/bin
	mkdir -p build-root/usr/sbin
	mkdir -p build-root/lib/orchestra/scores
	mkdir -p build-root/etc/conductor
	mkdir -p build-root/var/spool/orchestra
	install -m755 bin/conductor build-root/usr/sbin
	install -m755 bin/player build-root/usr/sbin
	install -m755 bin/submitjob build-root/usr/bin
	install -m755 bin/getstatus build-root/usr/bin
	install -m644 doc/examples/players build-root/etc/conductor

clean-build-root:
	-rm -rf build-root