FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /calculator-api ./cmd/api

FROM scratch
COPY --from=build /calculator-api /calculator-api
EXPOSE 8080
ENTRYPOINT ["/calculator-api"]
