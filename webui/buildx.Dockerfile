FROM node:26-alpine3.23

ENV WEBUI_DIR=/src/webui
RUN mkdir -p $WEBUI_DIR

COPY package.json yarn.lock .yarnrc.yml $WEBUI_DIR/

ENV VITE_APP_BASE_URL=""
ENV VITE_APP_BASE_API_URL="/api"

WORKDIR $WEBUI_DIR

RUN npm i -g corepack
RUN yarn workspaces focus --all --production

COPY . $WEBUI_DIR/

EXPOSE 8080
