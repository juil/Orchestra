include $(GOROOT)/src/Make.inc

#TARGDIR=../build/pkg/$(GOOS)_$(GOARCH)

TARG=orchestra
GOFILES=\
	orchestra.pb.go\
	wire.go\
	marshal.go\
	shared.go\
	messages.go\

include $(GOROOT)/src/Make.pkg
include $(GOROOT)/src/pkg/goprotobuf.googlecode.com/hg/Make.protobuf
