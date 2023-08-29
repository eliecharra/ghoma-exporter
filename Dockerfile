FROM golang:1.21 as build
WORKDIR /src
ADD go.mod go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 go build -o ghoma-exporter ./internal/cmd/server

FROM scratch
COPY --from=build /src/ghoma-exporter /bin/ghoma-exporter

EXPOSE      10005
ENTRYPOINT  [ "/bin/ghoma-exporter" ]
