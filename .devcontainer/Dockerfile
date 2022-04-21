# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.194.0/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version: 1, 1.16, 1.17
ARG VARIANT="latest"
FROM deis/go-dev:${VARIANT}

ENV GOPROXY=https://goproxy.cn,direct
ENV GOPRIVATE=github.com/Edgenesis,edgenesis.io/shifu
ENV GO111MODULE=on

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
     && apt-get -y install --no-install-recommends lsof apt-transport-https ca-certificates curl gnupg2 lsb-release \
     && curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | apt-key add - 2>/dev/null \
     && echo "deb [arch=amd64] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list \
     && apt-get update \
     && apt-get install -y docker-ce-cli

RUN curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64 \
     && chmod +x ./kind \
     && mv ./kind /usr/local/bin/kind

RUN curl -Lo ./kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH) \
     && chmod +x ./kubebuilder \ 
     && mv ./kubebuilder /usr/local/bin/kubebuilder

# [Optional] Uncomment the next line to use go get to install anything else you need
RUN go get -u github.com/ramya-rao-a/go-outline \
    k8s.io/apimachinery/pkg/util/wait

RUN go get golang.org/x/tools/gopls@latest \
    k8s.io/client-go@latest

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1