#set CPU architecture
FROM debian:latest
RUN apt-get update && apt-get upgrade -y
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y git golang-go

#Build Lattigo
WORKDIR /
RUN git clone https://github.com/tuneinsight/lattigo.git
WORKDIR /root/go/pkg/mod/github.com/tuneinsight/lattigo/
