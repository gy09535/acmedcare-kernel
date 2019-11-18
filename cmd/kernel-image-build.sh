go build kernel.go
docker build . -t kernel:0.1.0
docker tag kernel:0.1.0 docker.apiacmed.com/acmedback/kernel:0.1.0
docker push docker.apiacmed.com/acmedback/kernel:0.1.0
rm -rf kernel