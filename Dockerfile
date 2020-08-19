FROM mzz2017/git:alpine AS version
WORKDIR /build
ADD .git ./.git
RUN git describe --abbrev=0 --tags > ./version


FROM golang:alpine AS builder
ADD service /build/service
WORKDIR /build/service
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io
COPY --from=version /build/version ./
RUN export VERSION=$(cat ./version) && go build -ldflags="-X github.com/mzz2017/v2rayA/global.Version=${VERSION:1} -s -w" -o v2raya .

FROM node:lts-alpine AS builder-web
ADD gui /build/gui
WORKDIR /build/gui
RUN npm install --registry https://registry.npm.taobao.org && npm run build

FROM v2fly/v2fly-core AS v2ray

FROM bgiddings/iptables:latest
COPY --from=builder /build/service/v2raya /usr/bin/
COPY --from=builder-web /build/web /etc/v2raya/web
COPY --from=v2ray /usr/bin/v2ray/* /etc/v2ray/
ENV PATH=$PATH:/etc/v2ray
ENV GIN_MODE=release
EXPOSE 2017
ENTRYPOINT ["v2raya","--mode=universal"]

