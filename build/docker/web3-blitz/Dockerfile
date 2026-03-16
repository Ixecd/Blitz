FROM alpine:3.19

WORKDIR /app

COPY _output/platforms/linux/arm64/wallet-service /app/wallet-service

RUN chmod +x /app/wallet-service

EXPOSE 2113

CMD ["/app/wallet-service"]