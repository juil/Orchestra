include $(GOROOT)/src/Make.inc

TARGDIR=/usr/local/sbin
TARG=conductor
GCIMPORTS=-I../pkg/_obj
LDIMPORTS=-L../pkg/_obj

GOFILES=\
	conductor.go\
	dispatch.go\
	server.go\
	http.go\
	registry.go\
	client.go\
	signal.go\
	config.go\
	audience.go\

include $(GOROOT)/src/Make.cmd

testkey:	conductor_crt.pem conductor_key.pem

conductor_key.pem:
	openssl genrsa -out conductor_key.pem

HOSTNAME=$(shell hostname --fqdn)
conductor_crt.pem:	conductor_key.pem
	openssl req -new -x509 -key $< -outform pem -days 365 -out $@ -subj "/C=AU/ST=New South Wales/L=Sydney/CN=$(HOSTNAME)/"
