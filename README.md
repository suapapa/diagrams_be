# diagrams-back

## sandbox

Build and push docker image:

    $ cd sandbox
    $ docker build -f Dockerfile.gvisor -t diagrams_sandbox:dev .

Example Run:

    $ cat sample/k8s_diagram.py | docker run -i --rm diagrams_sandbox:dev

Explore container:

    $ docker run -it --rm -v $(pwd)/sample:/sample --entrypoint /bin/bash diagrams_sandbox:dev

## server

Make diagrams node json file:

    $ cd server
    $ make

Example Run:

    $ cd server
    $ go build && ./backend -l :8888
    $ cat ../sample/k8s_diagram.py | curl -X POST --data "$(</dev/stdin)" http://localhost:8888/diagram
