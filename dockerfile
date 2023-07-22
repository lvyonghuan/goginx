FROM golang as build
COPY . ./go/src/goginx
WORKDIR ./go/src/goginx
RUN go build -o goginxd

FROM ubuntu
COPY --from=build /go/go/src/goginx/goginxd .
CMD ["./goginxd"]