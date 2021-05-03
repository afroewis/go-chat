
FROM golang:1.16
WORKDIR /app
COPY ./*.go ./
COPY ./go.mod .
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM scratch
WORKDIR /app
COPY --from=0 /app/main/ ./
COPY public/home.html public/home.html
COPY public/css/style.css public/css/style.css
CMD ["/app/main"]
EXPOSE 8080
