# diagrams-back

## sandbox

Build and push docker image:

    $ cd sandox
    $ make

Example Run:

    $ cat sample/k8s_diagram.py | docker run -i --rm suapapa/diagrams-server-gvisor 

## server

Make diagrams node json file:

    $ cd server
    $ make

Example Run:

    $ cd server
    $ go build && ./backend -l :8888
    $ cat ../sample/k8s_diagram.py | curl -X POST --data "$(</dev/stdin)" http://localhost:8888/diagram
