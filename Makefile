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
GOPATH=$(PWD)/build-tree:$(PWD)
GOINSTALL_FLAGS=-dashboard=false -clean=true -u=false

export GOPATH


all: build

build:	build-tree
	goinstall $(GOINSTALL_FLAGS) conductor
	goinstall $(GOINSTALL_FLAGS) player
	goinstall $(GOINSTALL_FLAGS) submitjob
	goinstall $(GOINSTALL_FLAGS) getstatus

build-tree:
	mkdir -p build-tree/src
	mkdir -p build-tree/bin
	mkdir -p build-tree/pkg

clean:
	-$(RM) -r build-tree/pkg
	-$(RM) -r build-tree/bin

distclean:
	-$(RM) -r build-tree

deps:	distclean build-tree
	mkdir -p build-tree/src/github.com/kuroneko && cd build-tree/src/github.com/kuroneko && git clone http://github.com/kuroneko/configureit.git && cd configureit && git checkout v0.1
	mkdir -p build-tree/src/goprotobuf.googlecode.com/hg && cd build-tree/src/goprotobuf.googlecode.com/hg && hg clone -r release.r59 http://goprotobuf.googlecode.com/hg/proto

archive.release:
	git archive --format=tar --prefix=orchestra-$(VERSION)/ v$(VERSION) | gzip -9c > ../orchestra-$(VERSION).tgz

.PHONY : head.tgz

archive.head:
	git archive --format=tar --prefix=orchestra/ HEAD | gzip -9c > ../orchestra-HEAD.tgz
