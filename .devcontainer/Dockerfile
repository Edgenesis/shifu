# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.194.0/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version: 1, 1.16, 1.17
ARG VARIANT="latest"
FROM deis/go-dev:${VARIANT}

ENV GOPROXY=https://goproxy.io,direct
ENV GO111MODULE=on

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
     && apt-get -y install --no-install-recommends lsof

# [Optional] Uncomment the next line to use go get to install anything else you need
RUN go get -u github.com/ramya-rao-a/go-outline \
    k8s.io/apimachinery/pkg/util/wait
RUN GO111MODULE=on go get golang.org/x/tools/gopls@latest

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1