FROM node:6.3.0

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

RUN npm install -g gulp bower

COPY package.json $WEBUI_DIR/
COPY .bowerrc $WEBUI_DIR/
COPY bower.json $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN npm set progress=false
RUN npm install --quiet
RUN bower install --allow-root --quiet

COPY . $WEBUI_DIR/

EXPOSE 3000 3001 8080
