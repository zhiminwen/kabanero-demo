FROM node:lts-alpine
WORKDIR /app

COPY . /app
RUN npm install && npm run build

CMD ["node", "server/server.js"]

