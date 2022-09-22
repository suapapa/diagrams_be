# diagrams-back

## server

Build:

    $ go build

Example Run:

    $ ./diagram_srv

    $ cat sample/k8s_diagram.py | curl -X POST -d "$(</dev/stdin)" http://localhost:8080/diagram
