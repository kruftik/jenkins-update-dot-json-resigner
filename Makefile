CERT_PATH = ./cert

gen-cert:
	mkdir -p $(CERT_PATH)
	#openssl req -x509 \
#		-newkey rsa:4096 -keyout $(CERT_PATH)/your-update-center.key -nodes \
#		-sha256 -days 3650 -out $(CERT_PATH)/your-update-center.crt \
#	    -subj "/CN=YourJenkinsUpdateCenter"
	openssl genrsa -out $(CERT_PATH)/your-update-center.key 4096
	openssl req -new -key $(CERT_PATH)/your-update-center.key -out $(CERT_PATH)/your-update-center.csr -subj "/CN=YourJenkinsUpdateCenter"

	openssl x509 \
		-req -in $(CERT_PATH)/your-update-center.csr \
		-signkey $(CERT_PATH)/your-update-center.key \
		-sha256 -days 3650 -extfile $(CERT_PATH)/openssl-v3.ext \
		-out $(CERT_PATH)/your-update-center.crt

gen-test-cert:
	mkdir -p $(CERT_PATH)
	openssl req -x509 \
		-newkey rsa:4096 -keyout ./testdata/certs/test.key -nodes \
		-sha256 -days 3650 -out ./testdata/certs/test.crt \
	    -subj "/CN=YourJenkinsUpdateCenter"

test:
	go test ./... -count=1

build:
	go build -o app -v ./cmd

run-with-file:
	go run ./cmd/... \
		--debug \
		--certificate-path ./cert/your-update-center.crt \
		--key-path ./cert/your-update-center.key \
		--new-download-uri http://updates.jenkins.io/current/ \
		--listen-addr 127.0.0.1 \
		--listen-port 8282 \
		--update-json-path ./testdata/update-center/update-center.jsonp

run-with-url:
	go run ./cmd/... \
	  	--debug \
		--certificate-path ./cert/your-update-center.crt \
		--key-path ./cert/your-update-center.key \
		--new-download-uri http://updates.jenkins.io/current/ \
		--listen-addr 127.0.0.1 \
		--listen-port 8282 \
		--update-json-url http://updates.jenkins.io/update-center.json \
		--cache-ttl 30s

run-with-file-tls:
	go run ./cmd/... \
		--debug \
		--certificate-path ./cert/your-update-center.crt \
		--key-path ./cert/your-update-center.key \
		--new-download-uri http://updates.jenkins.io/current/ \
		--listen-addr 127.0.0.1 \
		--listen-port 8443 \
		--tlscert ./cert/your-update-center.crt \
		--tlskey ./cert/your-update-center.key \
		--update-json-path ./testdata/update-center/update-center.jsonp
