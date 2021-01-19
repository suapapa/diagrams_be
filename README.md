# diagrams-server

    $ docker build -t diagrams .
    $ docker run -it --rm -v $(pwd)/sample:/diagrams --runtime=runsc diagrams diagram.py
