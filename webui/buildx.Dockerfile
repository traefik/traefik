# syntax=docker/dockerfile:1.2
# Portal UI dependencies
FROM node:18-alpine AS dashboard-ui-deps

WORKDIR /app

COPY package.json yarn.lock ./

RUN yarn install --frozen-lockfile

# Portal UI build
FROM node:18-alpine AS dashboard-ui-builder

WORKDIR /app

COPY --from=dashboard-ui-deps /app/node_modules ./node_modules
COPY . .

ENV VITE_APP_BASE_URL ""
ENV VITE_APP_BASE_API_URL "/api"
ENV VITE_APP_DOCS_URL "https://doc.traefik.io/traefik/"
ENV VITE_APP_REPO_URL "https://github.com/traefik/traefik"

RUN HUSKY_SKIP_INSTALL=1 yarn build

FROM scratch AS dashboard-ui-export

COPY --from=dashboard-ui-builder /app/dist dist
COPY --from=dashboard-ui-deps /app/yarn.lock .
