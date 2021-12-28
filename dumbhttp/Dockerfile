FROM golang:alpine AS builder
WORKDIR /app
COPY ./ ./
RUN go build

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dumbhttp /app
COPY --from=builder /app/main.html /app/
EXPOSE 31323
CMD ["/app/dumbhttp"]
