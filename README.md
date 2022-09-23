# diagrams_srv

API server for the [diagrams](https://diagrams.mingrammer.com/)

## Test run

Build:

```bash
go build
```

Example Run:

```bash
./diagram_srv
```

It starts download docker image of the diagrams and node json.
Check its readiness via `/ready` endpoint.

Get diagram:

```bash
curl -X POST -d "$(cat sample/k8s_diagram.py)" http://localhost:8080/diagram
```

Get node info:

```bash
curl -X GET http://localhost:8080/nodes
```

## Test using docker

Build image:

```bash
docker build -t suapapa/diagrams_srv:dev .
docker push suapapa/diagrams_srv:dev
```

Run (only on linux):

```bash
sudo docker run -it --rm -v /var/run:/var/run -p 8080:8080 suapapa/diagrams_srv:dev
```