# diagrams-server

## sandbox

Build and push docker image:

    $ cd sandox
    $ make push

Run:

    $ cat sample/k8s_diagram.py | docker run -i --rm suapapa/diagrams-server-gvisor 

## backend

Make diagrams node json file:

    $ docker run -it --rm --entrypoint=/usr/local/bin/python suapapa/diagrams-server-gvisor:latest listup_nodes.py

Run:

    $ cd backend
    $ go build && ./backend -l :8888
    $ cat ../sample/k8s_diagram.py | curl -X POST --data "$(</dev/stdin)" http://localhost:8888/diagram