### build stage ###
FROM golang:1.16 as builder
# init setting
WORKDIR /workdir
ARG APP_VERSION
ARG APP_COMMIT
# download packages
COPY go.mod go.sum ./
RUN go mod download
# build
COPY . ./
RUN GOOS=linux go build -ldflags "-X main.appVersion=${APP_VERSION} -X main.appCommit=${APP_COMMIT}" .

### run stage ###
FROM gcr.io/distroless/base
COPY --from=builder /workdir/record-syncer .
ENTRYPOINT ["./record-syncer"]

