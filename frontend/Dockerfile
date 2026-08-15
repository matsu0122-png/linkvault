# --- build stage ---
FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# VITE_*はビルド時に静的ファイルへ埋め込まれる（実行時の環境変数では
# 変更できない）。ブラウザから見えるURLを渡すこと（例: http://localhost:8080）。
# compose.yamlのbackendサービス名（http://backend:8080）を渡してはいけない。
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
RUN npm run build

# --- runtime stage ---
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
