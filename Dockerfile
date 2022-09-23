FROM golang:1.19 as builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o app

FROM docker
WORKDIR /diagrams_srv
COPY --from=builder /src/app .
ENTRYPOINT ["/diagrams_srv/app"]
CMD [""]
