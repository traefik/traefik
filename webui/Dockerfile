FROM node:5.4

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

RUN npm install -g gulp bower

COPY package.json $WEBUI_DIR/
COPY .bowerrc $WEBUI_DIR/
COPY bower.json $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN npm install
RUN bower install --allow-root

COPY . $WEBUI_DIR/

EXPOSE 3000 3001 8080
