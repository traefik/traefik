FROM node:6.16.0

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

RUN apt-get -yq update \
&& apt-get -yq --no-install-suggests --no-install-recommends --force-yes install apt-transport-https \
&& curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
&& echo "deb https://dl.yarnpkg.com/debian/ stable main" |  tee /etc/apt/sources.list.d/yarn.list \
&& apt-get -yq update && apt-get -yq --no-install-suggests --no-install-recommends --force-yes install yarn \
&& rm -fr /var/lib/apt/lists/

COPY package.json $WEBUI_DIR/
COPY yarn.lock $WEBUI_DIR/

WORKDIR $WEBUI_DIR
RUN npm set progress=false
RUN yarn install

COPY . $WEBUI_DIR/

EXPOSE 8080
