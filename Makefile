all: linux-binaries windows-binaries

clean:
	rm -rf bin

ALL_GO_FILES := $(shell find ./ -name '*.go')

export CGO_ENABLED=0

.PHONY: linux-binaries
linux-binaries: bin/cboc-parse-reports bin/cboc-render bin/cboc-report

bin/cboc-parse-reports: $(ALL_GO_FILES)
	@mkdir -p bin
	go build -o $@ ./cmd/parse-reports/...

bin/cboc-render: $(ALL_GO_FILES)
	@mkdir -p bin
	go build -o $@ ./cmd/render/...

bin/cboc-report: $(ALL_GO_FILES)
	@mkdir -p bin
	go build -o $@ ./cmd/report/...

.PHONY: windows-binaries
windows-binaries: bin/cboc-parse-reports.exe bin/cboc-render.exe bin/cboc-report.exe

bin/cboc-parse-reports.exe: $(ALL_GO_FILES)
	@mkdir -p bin
	GOOS=windows go build -o $@ ./cmd/parse-reports/...

bin/cboc-render.exe: $(ALL_GO_FILES)
	@mkdir -p bin
	GOOS=windows go build -o $@ ./cmd/render/...

bin/cboc-report.exe: $(ALL_GO_FILES)
	@mkdir -p bin
	GOOS=windows go build -o $@ ./cmd/report/...
