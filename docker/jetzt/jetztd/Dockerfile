FROM golang:1.14 as build
WORKDIR /build
COPY cmd/jetztd/jetztd.go .
RUN go build jetztd.go

FROM debian:buster-slim
COPY --from=build /build/jetztd /jetztd
CMD ["/jetztd", "-target", "/data"]
EXPOSE 80
