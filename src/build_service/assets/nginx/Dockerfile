# Stage 1: Build certs
FROM alpine:3.19 AS certgen
RUN apk add --no-cache curl nss-tools ca-certificates openssl
RUN curl -L -o /usr/local/bin/mkcert https://github.com/FiloSottile/mkcert/releases/latest/download/mkcert-v1.4.4-linux-amd64 \
  && chmod +x /usr/local/bin/mkcert
WORKDIR /certs
ARG DOMAIN
RUN mkcert -install && mkcert $DOMAIN

# Stage 2: Final image
FROM nginx:alpine
ARG DOMAIN
COPY --from=certgen /certs/$DOMAIN.pem /etc/nginx/certs/cert.pem
COPY --from=certgen /certs/$DOMAIN-key.pem /etc/nginx/certs/key.pem
COPY default.conf /etc/nginx/conf.d/default.conf
EXPOSE 443
CMD ["nginx", "-g", "daemon off;"]
