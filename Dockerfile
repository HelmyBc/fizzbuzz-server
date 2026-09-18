# --- Build stage ---
FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache module downloads separately from source changes.
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

# --- Runtime stage ---
# Distroless has no shell, no package manager, and no unnecessary binaries,
# minimizing the attack surface of the production image.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
