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
cat sample/k8s_diagram.py | curl -X POST -d "$(</dev/stdin)" http://localhost:8080/diagram
```

Get node info:

```bash
curl -X GET http://localhost:8080/nodes
```

## Test using docker

Build image:

```bash
docker build -t diagrams_srv:dev .
```

Run (only on linux):

```bash
docker run -it --rm -v /var/run:/var/run -p 8080:8080 diagrams_srv:dev
```