FROM oven/bun:latest AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY package.json bun.lock ./
RUN bun install

# Copy the rest of the source code
COPY . .

# Build the frontend
ARG VITE_REACT_APP_SERVER_URL
RUN VITE_REACT_APP_SERVER_URL=${VITE_REACT_APP_SERVER_URL} bun run build

# Stage 2: Prepare the output
FROM alpine
COPY --from=builder /app/dist /dist
