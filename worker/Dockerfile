FROM node:12.16.2-alpine3.11

WORKDIR /src
COPY ./ /src/
RUN yarn build

CMD yarn -s start
