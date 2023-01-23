FROM golang:1.19 as builder
ARG PROGRAM_VER=dev-docker
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X main.programVer=${PROGRAM_VER}" -o app

FROM docker
WORKDIR /diagrams_be
COPY --from=builder /src/app .
ENTRYPOINT ["/diagrams_be/app"]
CMD [""]
