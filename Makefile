SOURCES=*.go

include Makefile.inc

build: bin/melp

bin/melp: $(SOURCES) Makefile go.mod go.sum
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
	bin/melp --port 9090 -l 8

docker:
	-docker rmi melp:docker
	docker build -t melp:docker .

nerdctl:
	-docker rmi melp:nerdctl
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

scan:
	trivy fs .

release:
	@echo -n "Latest release"
	@gh release list -L 1 | cat
	@echo ""
	gh release create

test/a:
	envexec tests/a_produce.env -- bin/melp --reconnect-delay 12s -f tests/a_produce.yaml --allow-stop

test/hdr:
	envexec tests/hdr_produce.env -- bin/melp -f tests/hdr_produce.yaml