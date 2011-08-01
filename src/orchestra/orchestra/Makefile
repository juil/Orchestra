include $(GOROOT)/src/Make.inc

#TARGDIR=../build/pkg/$(GOOS)_$(GOARCH)

TARG=orchestra
GOFILES=\
	orchestra.pb.go\
	wire.go\
	marshal.go\
	shared.go\
	request.go\
	registry.go\

include $(GOROOT)/src/Make.pkg

ifdef HAVE_PROTOBUF
include $(GOROOT)/src/pkg/goprotobuf.googlecode.com/hg/Make.protobuf
endif
