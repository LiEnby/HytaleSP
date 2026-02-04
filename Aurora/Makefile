CC = gcc
OUTDIR = Build
CFLAGS = -O3 -fPIC -fno-stack-protector -fshort-wchar -shared
OBJ = cs_string.o shared.o patch.o

ifeq ($(OS),Windows_NT)
    DLL := .dll
	DELCMD := del /Q /F /S
	OBJ += exports.o
else
    DLL := .so
	DELCMD := rm -rf
	CFLAGS += -nostdlib
endif

BINARY=$(OUTDIR)/Aurora$(DLL)

all: setup $(BINARY)


%.o: %.c $(DEPS)
	$(CC) -c -o $@ $< $(CFLAGS)

$(BINARY): $(OBJ)
	$(CC) -o $@ $^ $(CFLAGS)


.PHONY: setup
setup:
	-mkdir $(OUTDIR)

.PHONY: clean
clean:
	-$(DELCMD) *.o
	-$(DELCMD) Build
