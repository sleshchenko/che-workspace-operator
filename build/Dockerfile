FROM registry.access.redhat.com/devtools/go-toolset-rhel7:1.12.12-4 as builder

ENV PATH=/opt/rh/go-toolset-1.12/root/usr/bin:$PATH \
    GOPATH=/go/

USER root

RUN case $(uname -m) in \
       x86_64) ARCH="amd64" ;; \
       s390x) ARCH="s390x";; \
esac

WORKDIR /che-workspace-operator

COPY . .

# compile workspace controller binaries
RUN CGO_ENABLED=0 GOOS=linux go build \
  -mod=vendor \
  -o _output/bin/che-workspace-controller \
  -gcflags all=-trimpath=/ \
  -asmflags all=-trimpath=/ \
  cmd/manager/main.go

FROM registry.access.redhat.com/ubi8-minimal:8.1-279
COPY --from=builder /che-workspace-operator/_output/bin/che-workspace-controller /usr/local/bin/che-workspace-controller
COPY --from=builder /che-workspace-operator/internal-registry  internal-registry

ENV USER_UID=1001 \
    USER_NAME=che-workspace-controller

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/entrypoint"]
CMD /usr/local/bin/che-workspace-controller
