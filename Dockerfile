FROM golang:1.25-alpine AS build

WORKDIR /go/src/app
COPY . .

ENV CGO_ENABLED=0
RUN go mod download &&\
    go build -ldflags="-s -w" -o main .


FROM gcr.io/distroless/static:nonroot

COPY --from=build /go/src/app/main /

ARG IMAGE_TAG
ENV VERSION=${IMAGE_TAG}
ENV GIN_MODE=release
EXPOSE 8080
CMD ["/main"]
