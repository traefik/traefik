FROM node:12.18

ENV WEBUI_DIR /src/webui
ARG ARG_PLATFORM_URL=https://pilot.traefik.io
ENV PLATFORM_URL=${ARG_PLATFORM_URL}
RUN mkdir -p $WEBUI_DIR

COPY package.json $WEBUI_DIR/
COPY package-lock.json $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN npm install

COPY . $WEBUI_DIR/

EXPOSE 8080

RUN npm run lint
