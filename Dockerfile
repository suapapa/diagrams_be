FROM golang:1.19 as builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build "-X main.programVer=${PROGRAM_VER}" -o app

FROM docker
WORKDIR /diagrams_be
COPY --from=builder /src/app .
ENTRYPOINT ["/diagrams_be/app"]
CMD [""]
