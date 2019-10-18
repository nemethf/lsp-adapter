FROM golang:1.10-alpine AS lsp-adapter
WORKDIR /go/src/github.com/sourcegraph/lsp-adapter
COPY . .
#RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install github.com/sourcegraph/lsp-adapter
RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install github.com/nemethf/lsp-adapter
#RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install 

# 👀 Add steps here to build the language server itself 👀
# CMD ["echo", "🚨 This statement should be removed once you have added the logic to start up the language server! 🚨 Exiting..."]

FROM node:9

LABEL \
  maintainer="felician.nemeth@gmail.com" \
  Description="vscode-json-languageserver with lsp-adapter"

# Use tini as entrypoint to correctly handle signals and zombie processes.
ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]

USER node
# see https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md#global-npm-dependencies
ENV NPM_CONFIG_PREFIX=/home/node/.npm-global
ENV PATH=$PATH:/home/node/.npm-global/bin
ARG SERVER_VER=1.2.1
RUN npm install -g vscode-json-languageserver@$SERVER_VER

COPY --from=lsp-adapter /usr/local/bin/lsp-adapter /usr/local/bin
EXPOSE 8080
# Modify this command to connect to the language server
CMD ["lsp-adapter", "--trace", "--glob=*.json:*.jsonc", "--proxyAddress=0.0.0.0:8080", "vscode-json-languageserver", "--stdio"]
