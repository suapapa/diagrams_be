docker:
	docker build -t suapapa/diagrams-server-gvisor:latest -f ./Dockerfile.gvisor .

push: docker
	docker push suapapa/diagrams-server-gvisor:latest

.PHONY: docker push
