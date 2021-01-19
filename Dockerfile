FROM python:3
RUN apt update
RUN apt install -y graphviz
RUN pip install diagrams
WORKDIR /diagrams
ENTRYPOINT ["python"]
CMD [""]
