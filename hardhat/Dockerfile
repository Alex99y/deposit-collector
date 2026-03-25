FROM node:22-slim

WORKDIR /app

COPY hardhat/package.json .
COPY hardhat/package-lock.json .
COPY hardhat/tsconfig.json .
COPY hardhat/hardhat.config.ts .
COPY common ./common
COPY hardhat/start-hardhat.sh .
RUN chmod -R 777 start-hardhat.sh
RUN npm install
