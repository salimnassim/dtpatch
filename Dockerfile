FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /dtpatch ./cmd

FROM scratch

COPY --from=build /dtpatch /dtpatch

ENTRYPOINT ["/dtpatch"]
