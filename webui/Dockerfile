FROM node:6.9.1

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

COPY package.json $WEBUI_DIR/
COPY yarn.lock $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN npm set progress=false
RUN npm install --quiet --global yarn@0.16.1
RUN yarn install

COPY . $WEBUI_DIR/

EXPOSE 3000 3001 8080
