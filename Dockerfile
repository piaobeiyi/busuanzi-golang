FROM golang:alpine AS builder
WORKDIR /build
ADD . .
RUN  go build -o busuanzi-go .

FROM alpine
EXPOSE 18080
WORKDIR /
COPY --from=builder /build/busuanzi-go busuanzi-go
COPY --from=builder /build/busuanzi.js busuanzi.js
ENTRYPOINT ["/busuanzi-go"]
