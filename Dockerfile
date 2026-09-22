FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o schedule-app .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/schedule-app .
EXPOSE 8080
CMD ["./schedule-app"]
