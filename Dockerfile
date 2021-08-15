FROM docker.io/golang:1.20.2-alpine3.17 AS build_deps
RUN apk add --no-cache git

WORKDIR /workspace

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_deps AS build
COPY . .
RUN CGO_ENABLED=0 go build -o webhook -ldflags '-w -extldflags "-static"' .

FROM gcr.io/distroless/static-debian11:nonroot
COPY --from=build /workspace/webhook /
ENTRYPOINT ["/webhook"]
