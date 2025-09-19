# Sample Dockerfile for Image Builder Demo
FROM alpine:latest

# Add a simple label
LABEL maintainer="cnoe.io"
LABEL description="Sample application for image builder demo"

# Create app directory
WORKDIR /app

# Copy application files
COPY . .

# Install dependencies (simulate)
RUN echo "Installing dependencies..." && \
    echo "Application ready" > /app/status.txt

# Expose port
EXPOSE 8080

# Default command
CMD ["echo", "Hello from Image Builder Demo!"]