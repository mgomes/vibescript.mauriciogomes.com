FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/app ./

FROM alpine:3.22

RUN adduser -D -g "" app
USER app

COPY --from=build /bin/app /bin/app

CMD ["/bin/app"]
