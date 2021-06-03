FROM golang:1.15-alpine AS build
ADD . /app
ENV CGO_ENABLED 0
WORKDIR /app
RUN go build -o marketcap ./cmd

FROM scratch
COPY --from=build /app/marketcap /marketcap
CMD ["/marketcap"]
EXPOSE 9999