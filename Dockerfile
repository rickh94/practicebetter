FROM node:latest AS bunbuilder
WORKDIR /app
COPY package.json package-lock.json .
RUN npm install
COPY . .
RUN npm run build

FROM golang:1.21.5 AS gobuilder

WORKDIR /app
RUN mkdir -p /data/db
RUN go install github.com/a-h/templ/cmd/templ@v0.2.513
RUN ATLAS_VERSION=0.15.0 curl -sSf https://atlasgo.sh | sh -s -- --yes
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN rm -rf internal/static/dist
RUN mkdir -p internal/static/dist
COPY --from=bunbuilder /app/internal/static/dist/* ./internal/static/dist
RUN templ generate && \
CGO_ENABLED=1 GOOS=linux go build -o /practicebetter cmd/main.go

CMD touch $DB_PATH && atlas migrate apply --env prod && /practicebetter
