FROM golang:1.25 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /vanity \
    .

FROM scratch

COPY --from=build /vanity /usr/local/bin/vanity

EXPOSE 3030

ENTRYPOINT ["vanity"]

CMD ["serve", "-config", "/etc/vanity/vanity.yaml", "-listen", ":3030"]