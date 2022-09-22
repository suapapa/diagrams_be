# diagrams-back

## sandbox

Build and push docker image:

    $ docker build -f Dockerfile.gvisor -t diagrams_sandbox:dev .

Example Run:

    $ cat ../sample/k8s_diagram.py | docker run -i --rm diagrams_sandbox:dev

Explore container:

    $ docker run -it --rm -v $(pwd)/sample:/sample --entrypoint /bin/bash diagrams_sandbox:dev
