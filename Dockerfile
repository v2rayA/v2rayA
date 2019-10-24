FROM golang:1.13.2-stretch AS builder
ADD ./service /service
WORKDIR /service
RUN go build -mod=vendor -o V2RayA .

FROM mzz2017/v2ray-service
# https://github.com/mzz2017/docker-v2ray-service
WORKDIR /v2raya
COPY --from=builder /service/V2RayA /v2raya/
ENV GIN_MODE=release
EXPOSE 2017 1080-1082
ENTRYPOINT ["/v2raya/V2RayA"]
