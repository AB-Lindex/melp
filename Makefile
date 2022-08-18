SOURCES=*.go

include Makefile.inc

build: bin/melp

bin/melp: $(SOURCES) Makefile
	@mkdir -p bin
	go build -o bin/melp .

run: bin/melp
	ENDPOINT=$(ENDPOINT) \
	OUTPUT_KEY=$(OUTPUT_KEY) \
	OUTPUT_SECRET=$(OUTPUT_SECRET) \
	TOPIC=$(TOPIC) \
	OUTPUT_BEARER=$(OUTPUT_BEARER) \
	INPUT_KEY=$(INPUT_KEY) \
	INPUT_SECRET=$(INPUT_SECRET) \
	CONSUMERGROUP=$(CONSUMERGROUP) \
	bin/melp --port 9090

docker:
	docker build -t melp:docker .

nerdctl:
	nerdctl build -t melp:nerdctl .

check:
	@echo "Checking...\n"
	gocyclo -over 15 . || echo -n ""
	@echo ""
	golint -min_confidence 0.21 -set_exit_status ./...
	@echo ""
	go mod verify
	@echo "\nAll ok!"

check2:
	@echo ""
	golangci-lint run -E misspell -E depguard -E dupl -E goconst -E gocyclo -E ifshort -E predeclared -E tagliatelle -E errorlint -E godox -D structcheck
