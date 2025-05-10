FROM golang:alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS dev
WORKDIR /app
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air.toml"]

FROM base AS build
RUN go build -o /build/disco-backend

FROM gcr.io/distroless/static-debian12 AS prod
WORKDIR /app
COPY --from=build /build/disco-backend ./
CMD ["/app/disco-backend"]

