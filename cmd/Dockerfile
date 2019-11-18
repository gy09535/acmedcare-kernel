# cd $GOPATH
# docker build . -t kernel:0.1.0 -f ./src/acmed.com/kernel/Dockerfile
# docker tag kernel:0.1.0 docker.apiacmed.com/acmedback/kernel:0.1.0
# docker push docker.apiacmed.com/acmedback/kernel:0.1.0

FROM openjdk:8-jre-alpine
MAINTAINER Stefan.Gu
COPY . /go
RUN chmod -R 777 "/go"
