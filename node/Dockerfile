FROM node:alpine
WORKDIR /opt/app

RUN apk add --update --no-cache ca-certificates \
    && rm -rf /var/cache/apk/* \
    && addgroup -S nodeuser \
    && adduser -D -S -h /opt/app -s /sbin/nologin -G nodeuser nodeuser

RUN chmod g+rwx /opt/app

COPY package.json .
COPY package-lock.json .
COPY tsconfig.json .
COPY  ./build .
# For npm@5 or later, copy package-lock.json as well
# COPY package.json package-lock.json .

RUN npm install --only=production

RUN chown -R nodeuser:nodeuser /opt/app

EXPOSE  9097
RUN     addgroup nodeuser root
USER    nodeuser

CMD [ "npm", "run-script",  "start-docker" ]