FROM v2fly/v2fly-core AS v2ray
RUN ls /usr/local/share/v2ray || (mkdir -p /usr/local/share/v2ray && touch /usr/local/share/v2ray/.copykeep)

FROM golang:alpine
RUN apk --no-cache add iptables ip6tables git
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
RUN go get github.com/codegangsta/gin
WORKDIR /service
COPY --from=v2ray /usr/bin/v2ray /usr/bin/v2ctl /usr/share/v2ray/
COPY --from=v2ray /usr/local/share/v2ray/* /usr/local/share/v2ray/
ENV PATH=$PATH:/usr/share/v2ray
ENV CONFIG=../config.json
ENV GIN_BIN=../v2rayA
ENV GIN_GUILD_ARGS="-o ${GIN_BIN}"
EXPOSE 2017
ENTRYPOINT gin -a 2017 -i run