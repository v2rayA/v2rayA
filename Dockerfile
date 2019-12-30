FROM golang:alpine AS builder
ADD ./service /service
WORKDIR /service
ENV GOPROXY=https://goproxy.io
RUN go build -o V2RayA .


FROM alpine:latest  
RUN apk --no-cache add iptables
WORKDIR /service
COPY --from=builder /service/V2RayA .
ENV GIN_MODE=release
EXPOSE 2017
ENTRYPOINT "./V2RayA"
