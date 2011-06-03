include $(GOROOT)/src/Make.inc

TARGDIR=/usr/local/sbin
TARG=conductor

GOFILES=\
	conductor.go\
	dispatch.go\
	server.go\

testkey:	conductor_crt.pem conductor_key.pem

conductor_key.pem:
	openssl genrsa -out conductor_key.pem

HOSTNAME=$(shell hostname --fqdn)
conductor_crt.pem:	conductor_key.pem
	openssl req -new -key $< -out $@ -subj "/C=AU/ST=New South Wales/L=Sydney/CN=$(HOSTNAME)"


include $(GOROOT)/src/Make.cmd
