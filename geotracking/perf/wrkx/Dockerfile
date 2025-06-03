FROM ubuntu:latest as build
RUN apt update
RUN apt install -y gcc make git libssl-dev zlib1g-dev
RUN git clone https://github.com/vbogretsov/wrkx.git

WORKDIR /wrkx
RUN git checkout luajit-2.1
RUN make

FROM ubuntu:latest
COPY --from=build /wrkx/wrk /usr/local/bin/wrk
ENTRYPOINT [ "wrk" ]
