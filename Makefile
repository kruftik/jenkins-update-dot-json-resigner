CERT_PATH = ./cert

gen-cert:
	mkdir -p $(CERT_PATH)
	openssl req -x509 \
		-newkey rsa:4096 -keyout $(CERT_PATH)/your-update-center.key -nodes \
		-sha256 -days 3650 -out $(CERT_PATH)/your-update-center.crt \
	    -subj "/CN=YourJenkinsUpdateCenter"

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
