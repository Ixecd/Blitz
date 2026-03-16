FROM BASE_IMAGE

WORKDIR /app

COPY wallet-service /app/wallet-service

RUN chmod +x /app/wallet-service

EXPOSE 2113

CMD ["/app/wallet-service"]