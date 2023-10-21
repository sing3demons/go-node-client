FROM node:16.17.0-alpine
RUN apk update && apk add git
ENV TZ=Asia/Bangkok
WORKDIR /usr/src/app


COPY package*.json ./
RUN npm install
COPY . .

EXPOSE 3000
CMD ["node","index.js"]