build:
	GOOS=js GOARCH=wasm go build -o main.wasm .
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" .

serve: build
	python3 -m http.server 8080

clean:
	rm -f main.wasm wasm_exec.js

.PHONY: build serve clean
