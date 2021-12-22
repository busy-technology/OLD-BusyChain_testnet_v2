FROM node:14

# Create app directory
WORKDIR /usr/src/app

# Install app dependencies
# A wildcard is used to ensure both package.json AND package-lock.json are copied
# where available (npm@5+)
COPY package*.json ./

# If you are building your code for production
# RUN npm installRUN npm ci --only=production

# Bundle app source
COPY . .
RUN rm -rf node_modules

RUN npm install

EXPOSE 3000
CMD [ "node", "index.js" ]