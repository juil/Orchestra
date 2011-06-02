include $(GOROOT)/src/Make.inc

all: build

DIRS=\
	pkg\
	conductor\
	player\

clean.dirs: $(addsuffix .clean, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
nuke.dirs: $(addsuffix .nuke, $(DIRS))
build.dirs: $(addsuffix .build, $(DIRS))

%.clean:
	+$(MAKE) -C $* clean

%.install:
	+@echo install $*
	+@$(MAKE) -C $* install.clean >$*/build.out 2>&1 || (echo INSTALL FAIL $*; cat $*/build.out; exit 1)

%.nuke:
	+$(MAKE) -C $* nuke

%.build:
	+$(MAKE) -C $*

clean: clean.dirs

install: install.dirs

test:   test.dirs

build:	build.dirs
