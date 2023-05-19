# 编译
FROM golang as builder
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
RUN --mount=type=bind,source=.,destination=/chrome_docker \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd /chrome_docker && CGO_ENABLED=0 go build -ldflags='-s -w -extldflags "-static"' -o /tmp/chrome_service


FROM chromedp/headless-shell as app
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
RUN --mount=type=cache,sharing=locked,target=/var/lib/apt \
    --mount=type=cache,sharing=locked,target=/var/cache/apt \
    apt-get update && apt-get install -y --no-install-recommends \
        ttf-wqy-zenhei
COPY --from=builder /tmp/chrome_service /chrome_service
WORKDIR /

# Add Tini, to kill zombie process
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]
CMD [ "/chrome_service", "-addr", ":5558" ]