# Go API用の本番イメージ。api/migrate/createuser のバイナリを含み、
# docker-compose側でどれを起動するかをentrypoint指定で切り替える。

FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/createuser ./cmd/createuser

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/migrate /usr/local/bin/migrate
COPY --from=build /out/createuser /usr/local/bin/createuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
