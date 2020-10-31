FROM golang:alpine as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN go build -o bbgo  ./cmd/bbgo

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/bbgo /app/
WORKDIR /app
CMD ["./bbgo"]
