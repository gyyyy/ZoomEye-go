ver?=dev
NAME=ZoomEye-go_$(ver)
BINDIR=bin
GOBUILD=CGO_ENABLED=0 go build -ldflags '-w -s'

all: linux macos win64 win32

linux:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)_$@

macos:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)_$@

win64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)_$@.exe

win32:
	GOARCH=386 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)_$@.exe

release: linux macos win64 win32
	chmod +x $(BINDIR)/$(NAME)_*
	gzip $(BINDIR)/$(NAME)_linux
	gzip $(BINDIR)/$(NAME)_macos
	zip -m -j $(BINDIR)/$(NAME)_win32.zip $(BINDIR)/$(NAME)_win32.exe
	zip -m -j $(BINDIR)/$(NAME)_win64.zip $(BINDIR)/$(NAME)_win64.exe

clean:
	rm $(BINDIR)/*