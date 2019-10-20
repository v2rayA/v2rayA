FROM v2ray/official
LABEL maintainer="m@mzz.pub"
ADD . /service
WORKDIR /service
RUN go build -mod=vendor
EXPOSE 8080
ENTRYPOINT ["v2rayA"]