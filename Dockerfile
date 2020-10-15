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
RUN echo "network-timeout 600000" >> .yarnrc
#RUN yarn config set registry https://registry.npm.taobao.org
#RUN yarn config set sass_binary_site https://cdn.npm.taobao.org/dist/node-sass -g
RUN yarn
RUN yarn build

FROM v2fly/v2fly-core AS v2ray
RUN ls /usr/local/share/v2ray || (mkdir -p /usr/local/share/v2ray && touch /usr/local/share/v2ray/.copykeep)

FROM bgiddings/iptables:latest
COPY --from=builder /build/service/v2raya /usr/bin/
COPY --from=builder-web /build/web /etc/v2raya-web
COPY --from=v2ray /usr/bin/v2ray /usr/bin/v2ctl /usr/share/v2ray/
COPY --from=v2ray /usr/bin/*.dat /usr/local/share/v2ray/
COPY --from=v2ray /usr/local/share/v2ray/* /usr/local/share/v2ray/
ENV PATH=$PATH:/usr/share/v2ray
ENV GIN_MODE=release
EXPOSE 2017
VOLUME /etc/v2raya
ENTRYPOINT ["v2raya","--mode=universal", "--webdir=/etc/v2raya-web"]

