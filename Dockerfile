FROM golang:1.23 AS build-stage
WORKDIR /app
COPY . ./
WORKDIR /app/cmd
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /file-storage

FROM alpine AS release-stage
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=build-stage /file-storage ./file-storage

ENTRYPOINT ./file-storage --conf ./config.yml --folder ./data