FROM --platform=linux/amd64 registry.access.redhat.com/ubi8/ubi-minimal:8.6 AS builder

RUN microdnf install -y tar gzip make which
ENV ARCHITECTURE=x64
# install go 1.17.8
RUN curl -O -J https://dl.google.com/go/go1.18.4.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.18.4.linux-amd64.tar.gz
RUN ln -s /usr/local/go/bin/go /usr/local/bin/go

WORKDIR /workspace

COPY . ./

RUN go mod vendor 
RUN make binary

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.6

COPY --from=builder /workspace/authz /usr/local/bin/

EXPOSE 8000

ENTRYPOINT ["/usr/local/bin/authz","serve"]

LABEL name="authz" \
      version="0.0.1" \
      summary="authz service" \
      description="authz"
