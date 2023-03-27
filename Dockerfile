FROM registry.access.redhat.com/ubi9/ubi-minimal:9.1 AS builder
ARG TARGETARCH
USER root
RUN microdnf install -y tar gzip make which

# install platform specific go version
RUN curl -O -J  https://dl.google.com/go/go1.19.6.linux-${TARGETARCH}.tar.gz
RUN tar -C /usr/local -xzf go1.19.6.linux-${TARGETARCH}.tar.gz
RUN ln -s /usr/local/go/bin/go /usr/local/bin/go

WORKDIR /workspace

COPY . ./

RUN go mod vendor 
RUN make binary

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.1

COPY --from=builder /workspace/authz /usr/local/bin/
RUN mkdir -p /schema
COPY --from=builder /workspace/schema/spicedb_bootstrap.yaml  /schema/

EXPOSE 8080

USER 1001
ENTRYPOINT ["/usr/local/bin/authz","serve"]

LABEL name="authz" \
      version="0.0.1" \
      summary="authz service" \
      description="authz"
