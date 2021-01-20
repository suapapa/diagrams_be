# diagrams-server

## sandbox

Build and push docker image:

    $ cd sandox
    $ make push

Run:

    $ cat sample/k8s_diagram.py | docker run -i --rm suapapa/diagrams-server-gvisor 

## backend

Run:

    $ cd backend
    $ $ cat ../sample/k8s_diagram.py | curl -X POST --data "$(</dev/stdin)" http://localhost:8080/diagram