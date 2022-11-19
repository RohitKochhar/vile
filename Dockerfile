# Stage 1: Builder
FROM golang:1.16@sha256:d388153691a825844ebb3586dd04d1c60a2215522cc445701424205dffc8a83e as build

COPY ./src /src

WORKDIR /src


RUN CGO_ENABLED=0 GOOS=linux go build -o vile-server

# Stage 2: Runner
FROM scratch

COPY --from=build /src/vile-server .

COPY --from=build /src/keys/ /keys/

EXPOSE 8080

CMD ["/vile-server"]
