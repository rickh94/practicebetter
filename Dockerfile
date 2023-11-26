FROM oven/bun:latest AS bunbuilder
WORKDIR /app
COPY package.json bun.lockb .
RUN bun install
COPY . .
RUN bun run build

FROM golang:1.21 AS gobuilder

WORKDIR /app
RUN mkdir -p /data/db
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN curl -sSf https://atlasgo.sh | sh -s -- --yes
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN rm -rf internal/static/dist
RUN mkdir -p internal/static/dist
COPY --from=bunbuilder /app/internal/static/css/main.css ./internal/static/css/main.css
COPY --from=bunbuilder /app/internal/static/dist/* ./internal/static/dist
RUN templ generate
RUN CGO_ENABLED=1 GOOS=linux go build -o /practicebetter cmd/main.go

CMD touch $DB_PATH && atlas migrate apply --env prod && /practicebetter
