FROM golang:alpine
ADD ./service /service
WORKDIR /service
RUN go build -o V2RayA .
ENV GIN_MODE=release
EXPOSE 2017 20170-20172
ENTRYPOINT "./V2RayA"
