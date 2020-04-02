FROM node:13.12.0-alpine3.11

WORKDIR /home/src
COPY worker/ /home/src/
RUN yarn build

CMD yarn -s start
