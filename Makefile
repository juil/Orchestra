#
# version of Orchestra
#
VERSION=0.2.0

#
# packaging revision.
#
REVISION=1

# remove at your own peril.
#
# This tells goinstall to work against the local directory as the
# build/source path, and not use the system directories.
#
GOPATH=$(shell pwd)
GOINSTALL_FLAGS=-dashboard=false -clean=true -u=true

export GOPATH


all: build

build:
	goinstall $(GOINSTALL_FLAGS) conductor
	goinstall $(GOINSTALL_FLAGS) player
	goinstall $(GOINSTALL_FLAGS) submitjob
	goinstall $(GOINSTALL_FLAGS) getstatus

tgz:
	git archive --format=tar --prefix=orchestra-$(VERSION)/ v$(VERSION) | gzip -9c > ../orchestra-$(VERSION).tgz

deb:	build-root
	fpm -s dir -t deb \
		-n 'orchestra' \
		-v "$(VERSION)" \
		--iteration "$(REVISION)" \
		-m 'Chris Collins <chris.collins@anchor.net.au>' \
		-p "../orchestra_$(VERSION)-$(REVISION)_$(shell dpkg --print-architecture).deb" \
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
		--iteration "$(REVISION)" \
		-m 'Chris Collins <chris.collins@anchor.net.au>' \
		-p "../orchestra-$(VERSION)-$(REVISION).rpm" \
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
