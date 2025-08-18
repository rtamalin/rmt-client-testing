#
# Build Settings
#
ARG TESTER_BUILD_DIR=/var/cache/build/

#
# Go Settings
#
ARG GO_NO_PROXY=github.com/SUSE
ARG GOLANG_BASE=registry.suse.com/bci/golang
ARG GOLANG_VERSION=1.24-openssl

#
# SLE Verion Settings
ARG SLE_BCI_REG=registry.suse.com/bci
ARG SLE_BCI_IMAGE=bci-base
ARG SLE_BCI_VERSION=15.7

# suseconnect Settings
ARG CONNECT_REPO=SUSE/connect-ng
ARG CONNECT_REF=next

#
# Build the code in BCI golang based image
#
FROM ${GOLANG_BASE}:${GOLANG_VERSION} AS builder

# Reference top-level args used by this image
ARG CONNECT_REPO
ARG CONNECT_REF
ARG TESTER_BUILD_DIR
ARG GO_NO_PROXY

# create a temporary workspace
WORKDIR ${TESTER_BUILD_DIR}

# ensure required tools are available
RUN set -euo pipefail; \
        zypper -n \
          install --no-recommends \
          curl \
          git \
          make \
          ; \
        zypper -n clean

# retrieve the branch or tag hash before attempting to clone the repo to
# ensure repo ref is both valid and, if it exists, will be freshly cloned
# if it's content changes, and then clone the repo.
RUN \
        reporef=.repo_ref.json; \
        if curl --silent --fail \
                https://api.github.com/repos/${CONNECT_REPO}/git/refs/heads/${CONNECT_REF} \
                > ${reporef}; then \
                echo Saved branch hash data; \
        elif curl --silent --fail \
                https://api.github.com/repos/${CONNECT_REPO}/git/refs/tags/${CONNECT_REF} \
                > ${reporef}; then \
                echo Saved tag hash data; \
        else \
                echo ${CONNECT_REF} is neither a branch or a tag in ${CONNECT_REPO}; \
        fi; \
        [[ -s ${reporef} ]] || exit 1; \
        git clone \
          --branch ${CONNECT_REF} \
          https://github.com/SUSE/connect-ng \
          connect-ng

RUN cd connect-ng; \
    export GONOPROXY=${GO_NO_PROXY} ; \
    make vendor && \
    make build

# copy in local sources other than the Dockerfile and container build
# related items, with less frequently changing items coming first
RUN mkdir -p rmt-client-testing/
COPY LICENSE README.md rmt-client-testing/
COPY Makefile Makefile.golang rmt-client-testing/
COPY go.mod go.sum rmt-client-testing/
COPY cmd rmt-client-testing/cmd/
COPY internal rmt-client-testing/internal/

# build the local sources within the container
RUN cd rmt-client-testing; \
    export GONOPROXY=${GO_NO_PROXY} ; \
    make mod-download build

#
# Create the rmt-client-tester image.
#

FROM ${SLE_BCI_REG}/${SLE_BCI_IMAGE}:${SLE_BCI_VERSION} AS rmt-client-tester

ARG TESTER_BUILD_DIR

# ensure required tools are available
RUN set -euo pipefail; \
        zypper -n \
          install --no-recommends \
          suseconnect-ng \
          pciutils \
          ; \
        zypper -n clean

# create the /app/bin directory and copy in built binaries
RUN mkdir -p /app/bin

# copy built binaries from builder image
COPY --from=builder --chmod=0755 ${TESTER_BUILD_DIR}/connect-ng/out/. /app/bin/
COPY --from=builder --chmod=0755 ${TESTER_BUILD_DIR}/rmt-client-testing/cmd/rmt-hwinfo-tester /app/bin/

COPY --chmod=0755 entrypoint.bash /app/

# setup the environment variables
#ENTRYPOINT ["/app/entrypoint.bash"]
#CMD ["help"]
