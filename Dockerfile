FROM golang:buster
ADD ./service /service
WORKDIR /service
RUN go build -o V2RayA .
RUN echo "deb http://mirrors.ustc.edu.cn/debian buster main contrib non-free" > /etc/apt/sources.list
RUN echo "deb http://mirrors.ustc.edu.cn/debian buster-updates main contrib non-free" >> /etc/apt/sources.list
RUN apt-get update
RUN apt-get install -y iptables
ENV GIN_MODE=release
EXPOSE 2017
ENTRYPOINT "./V2RayA"
