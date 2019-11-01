FROM golang:1.10-alpine AS lsp-adapter
WORKDIR /go/src/github.com/sourcegraph/lsp-adapter
COPY . .
RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install github.com/sourcegraph/lsp-adapter

FROM alpine:3.10 AS p4ls-build

RUN apk update && \
    apk add cmake boost boost-dev boost-static ninja \
            rapidjson rapidjson-dev gcc g++


ENV P4LS_COMMIT 54dd19d88cd561061ac3103dc452d6854e5899fa
ADD https://github.com/dmakarov/p4ls/archive/${P4LS_COMMIT}.zip /p4ls.zip
RUN unzip p4ls.zip && mv p4ls-${P4LS_COMMIT} p4ls && rm p4ls.zip

WORKDIR p4ls

RUN echo -e \
  "SET(CMAKE_FIND_LIBRARY_SUFFIXES \".a\") \n" \
  "SET(BUILD_SHARED_LIBS OFF) \n" \
  "SET(CMAKE_EXE_LINKER_FLAGS \"-static\")" > _static
RUN cat _static CMakeLists.txt > _c.txt && mv _c.txt CMakeLists.txt
RUN cat CMakeLists.txt

# poor man's patch
# remove gcc-specific debug code
RUN sed -i server/daemon/main.cpp -e '55,72d' -e '25d'

RUN mkdir -p build/ninja/release
WORKDIR build/ninja/release
RUN cmake -G Ninja -DCMAKE_BUILD_TYPE=Release \
    -DHUNTER_ENABLED=OFF -DUNITTESTS_ENABLED=OFF \
    /p4ls
RUN cmake --build . --target server/daemon/install

## I wasn't able to build p4c in alpine
#
# RUN apk add gc gc-dev g++ git autoconf automake cmake bison flex \
#             protobuf protobuf-dev gmp gmp-dev boost boost-static \
#             boost-dev make
#
# RUN git clone --recursive https://github.com/p4lang/p4c.git
# cd p4c
# mkdir build
# cd build
# cmake .. -DENABLE_EBPF=OFF -DENABLE_P4C_GRAPHS=OFF -DENABLE_P4TEST=OFF \
#          -DENABLE_GTESTS=OFF
# make

FROM debian:buster-slim AS p4c

RUN apt-get update && \
    apt-get install -y \
        cmake g++ git automake libtool libgc-dev bison flex libfl-dev \
        libgmp-dev libboost-dev libboost-iostreams-dev libboost-graph-dev \
        pkg-config python libprotobuf-dev protobuf-compiler && \
    apt-get clean

RUN git clone --depth=1 --recursive https://github.com/p4lang/p4c.git

WORKDIR p4c
RUN mkdir build
WORKDIR build
RUN cmake .. -DENABLE_EBPF=OFF -DENABLE_P4C_GRAPHS=OFF \
             -DENABLE_P4TEST=OFF -DENABLE_GTESTS=OFF
RUN make && make install

FROM debian:buster-slim

LABEL \
  maintainer="felician.nemeth@gmail.com" \
  Description="p4ls with lsp-adapter"

# # Use tini as entrypoint to correctly handle signals and zombie processes.
# ENV TINI_VERSION v0.18.0
# ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
# RUN chmod +x /tini
# ENTRYPOINT ["/tini", "--"]

RUN apt-get update \
    && apt-get install -y \
           python gcc libboost-iostreams1.67.0 libboost-regex1.67.0 \
           libgmpxx4ldbl libgc1c2 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=p4ls-build /usr/local/bin/p4lsd /usr/local/bin
COPY --from=lsp-adapter /usr/local/bin/lsp-adapter /usr/local/bin
COPY --from=p4c /usr/local/bin/p4* /usr/local/bin/
COPY --from=p4c /usr/local/share/p4c /usr/local/share
COPY dockerfiles/p4/lsp-adapter-replace /

#ENTRYPOINT ["/usr/local/bin/lsp-adapter"]
EXPOSE 8080
CMD ["/usr/local/bin/lsp-adapter", \
        "--trace", \
        "--glob=*.p4:compile_commands.json", \
        "--proxyAddress=0.0.0.0:8080", \
        "--beforeInitializeHook=/lsp-adapter-replace", \
        "/usr/local/bin/p4lsd", "-d"]
