# Stage 1: Testing (Runs unit tests)
FROM golang:1.16@sha256:d388153691a825844ebb3586dd04d1c60a2215522cc445701424205dffc8a83e as testing

WORKDIR /src

COPY  ./src/go.mod ./src/go.sum  ./

RUN go mod download

COPY ./src /src

RUN go test -v ./core ./server ./transaction_logs

# Stage 2: Builder (Builds executable)
FROM golang:1.16@sha256:d388153691a825844ebb3586dd04d1c60a2215522cc445701424205dffc8a83e as build

WORKDIR /src

COPY  ./src/go.mod ./src/go.sum  ./

RUN go mod download

COPY --from=testing /src /src

RUN CGO_ENABLED=0 GOOS=linux go build -o vile-server

# Stage 3: Runner
FROM scratch

COPY --from=build /src/vile-server .

COPY --from=build /src/keys/ /keys/

EXPOSE 8080

CMD ["/vile-server"]
