# Stage 1: Build React app
FROM node:20-alpine AS build-env

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# Stage 2: Serve with Nginx
FROM nginx:stable-alpine

# Remove default Nginx static content
RUN rm -rf /usr/share/nginx/html/*

# Copy React build output to Nginx html directory
COPY --from=build-env /app/build/client /usr/share/nginx/html
COPY --from=build-env /app/default.conf /etc/nginx/conf.d/default.conf

# Copy custom Nginx config if needed (optional)
# COPY nginx.conf /etc/nginx/nginx.conf

# Expose port 80 (default for Nginx)
EXPOSE 80

# Start Nginx in the foreground
CMD ["nginx", "-g", "daemon off;"]
