FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/smistudy-api . \
 && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/smistudy-api /smistudy-api
# Owned by the nonroot user so a fresh named volume mounted here is writable.
COPY --from=build --chown=nonroot:nonroot /out/data /data
ENV SMISTUDY_ADDR=:8080 \
    SMISTUDY_DB=/data/smistudy.db
EXPOSE 8080
VOLUME /data
ENTRYPOINT ["/smistudy-api"]
