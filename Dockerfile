FROM node:stretch AS build-node

WORKDIR /app

COPY web/package*.json ./
RUN npm ci --quiet 

COPY web/ .
RUN npx parcel build index.html --no-source-maps

FROM golang:1.15 AS build-go
 
WORKDIR /src
COPY go.* ./
RUN go mod download
 
COPY . .
RUN go build -o /go/bin/server github.com/wietsevenema/todo/cmd/server

FROM gcr.io/distroless/base-debian10:nonroot AS run

COPY --from=build-go /go/bin/server /app/server
COPY --from=build-node /app/dist /app/dist 
 
ENTRYPOINT ["/app/server"]
