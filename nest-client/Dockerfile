FROM node:16.17.0-alpine
RUN apk update && apk add git
ENV TZ=Asia/Bangkok
WORKDIR /usr/src/app


COPY package*.json ./
RUN npm install
COPY . .

RUN npm run build
RUN rm -rf src
EXPOSE 3000
CMD ["npm", "run", "start:prod"]