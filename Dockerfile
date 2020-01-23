FROM golang:latest as build
RUN mkdir /src
ADD src/ /src/
WORKDIR /src
RUN go build -o app main/main.go

FROM golang:latest
MAINTAINER Daniel Richter

COPY --from=build /src/app /usr/local/bin/app
CMD ["/usr/local/bin/app"]