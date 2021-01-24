all: json
	go build

json:
	docker run --rm --entrypoint=/usr/local/bin/python \
		suapapa/diagrams-server-gvisor:latest \
		listup_nodes.py > diagrams_nodes.json

.PHONY: all json
