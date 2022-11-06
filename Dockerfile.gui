FROM node:lts-alpine AS builder
ADD gui /gui
WORKDIR /gui
RUN export NODE_OPTIONS=--openssl-legacy-provider && yarn && yarn build

FROM nginx:stable-alpine
COPY --from=builder /web /usr/share/nginx/html
EXPOSE 80
