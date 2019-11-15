# cd $GOPATH
# docker build . -t kernel:0.1.0 -f ./src/acmed.com/kernel/Dockerfile
# docker tag kernel:0.1.0 docker.apiacmed.com/acmedback/kernel:0.1.0
# docker push docker.apiacmed.com/acmedback/kernel:0.1.0
FROM docker.apiacmed.com/acmedback/golang:1.12.13
COPY . /go