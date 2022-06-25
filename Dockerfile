# 编译
FROM golang as builder
RUN --mount=type=bind,source=.,destination=/chrome_docker \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd /chrome_docker && CGO_ENABLED=0 go build -ldflags='-s -w -extldflags "-static"' -o /tmp/chrome_service


FROM chromedp/headless-shell as app
RUN --mount=type=cache,sharing=locked,target=/var/lib/apt \
    --mount=type=cache,sharing=locked,target=/var/cache/apt \
    apt-get update && apt-get install -y --no-install-recommends \
        ttf-wqy-zenhei
COPY --from=builder /tmp/chrome_service /chrome_service
WORKDIR /
ENTRYPOINT [ "/chrome_service", "-addr", ":5558" ]