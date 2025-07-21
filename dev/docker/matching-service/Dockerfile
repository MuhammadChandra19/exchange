ARG GO_VERSION=1.24-alpine
FROM golang:${GO_VERSION}
RUN wget -O /bin/air https://github.com/cosmtrek/air/releases/download/v1.40.4/air_1.40.4_linux_amd64 && chmod +x /bin/air
WORKDIR /app
COPY . .
CMD ["/bin/air", "-c", "dev/docker/matching-service/.air.toml"]
