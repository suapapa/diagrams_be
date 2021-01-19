# diagrams-server

Build docker image:

    $ cd server
    $ docker build -t diagrams_server -f ./Dockerfile.gvisor .

Run:

    $ cat sample/k8s_diagram.py | docker run -i --rm --runtime=runsc diagrams_server:latest