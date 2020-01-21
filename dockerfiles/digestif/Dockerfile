# FROM golang:1.10-alpine AS lsp-adapter
# WORKDIR /go/src/github.com/sourcegraph/lsp-adapter
# COPY . .
# RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install github.com/sourcegraph/lsp-adapter

FROM debian:bullseye

# Use tini as entrypoint to correctly handle signals and zombie processes.
ENV TINI_VERSION v0.18.0
# ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
COPY --from=nemethf/vscode-json-languageserver /tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]

RUN apt-get update \
 && apt-get install -y \
        lua5.3 lua5.3-dev luarocks \
 && rm -rf /var/lib/apt/lists/*

COPY --from=nemethf/p4ls /usr/local/bin/lsp-adapter /usr/local/bin

RUN luarocks install digestif

EXPOSE 8080
CMD ["lsp-adapter", \
       "--trace", \
       "--proxyAddress=0.0.0.0:8080", \
       "digestif" \
    ]

#       "--glob=*.cpp:*.h:*.c:.ccls:compile_commands.json", \

