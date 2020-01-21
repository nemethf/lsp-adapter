FROM golang:1.10-alpine AS lsp-adapter
WORKDIR /go/src/github.com/sourcegraph/lsp-adapter
COPY . .
RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install github.com/sourcegraph/lsp-adapter

FROM debian:bullseye

# Use tini as entrypoint to correctly handle signals and zombie processes.
COPY --from=nemethf/vscode-json-languageserver /tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]

#COPY --from=nemethf/p4ls /usr/local/bin/lsp-adapter /usr/local/bin
COPY --from=lsp-adapter /usr/local/bin/lsp-adapter /usr/local/bin

RUN apt-get update \
 && apt-get install -y \
        clangd-9 \
 && rm -rf /var/lib/apt/lists/*

#COPY dockerfiles/ccls/lsp-adapter-replace /

EXPOSE 8080
CMD ["lsp-adapter", \
       "--trace", \
       "--proxyAddress=0.0.0.0:8080", \
       "-didOpenLanguage=cpp", \
       "clangd-9", \
       "-log=verbose" \
    ]
