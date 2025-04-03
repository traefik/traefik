FROM node:22.9-alpine3.20

ENV WEBUI_DIR=/src/webui
RUN mkdir -p $WEBUI_DIR

COPY package.json yarn.lock .yarnrc.yml $WEBUI_DIR/

WORKDIR $WEBUI_DIR

RUN corepack enable
RUN yarn workspaces focus --all --production

COPY . $WEBUI_DIR/

EXPOSE 8080

# RUN yarn lint