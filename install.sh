#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

echo -e "${GREEN}Installing FastHTTP for RHEL...${NC}"

# Detect RHEL version
if [ -f /etc/redhat-release ]; then
    RHEL_VERSION=$(cat /etc/redhat-release)
    echo -e "${YELLOW}Detected: $RHEL_VERSION${NC}"
else
    echo -e "${YELLOW}Warning: This script is designed for RHEL/CentOS/Rocky Linux${NC}"
fi

# Install required packages
echo -e "${GREEN}Installing dependencies...${NC}"
dnf install -y golang php-fpm || yum install -y golang php-fpm

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go installation failed. Please install Go manually.${NC}"
    exit 1
fi

# Get Go version
GO_VERSION=$(go version)
echo -e "${GREEN}Found: $GO_VERSION${NC}"

# Build the application
echo -e "${GREEN}Building FastHTTP...${NC}"
go build -o fasthttp fasthttp.go

if [ ! -f "./fasthttp" ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

# Install binary
echo -e "${GREEN}Installing binary to /usr/local/bin/fasthttp...${NC}"
cp fasthttp /usr/local/bin/fasthttp
chmod +x /usr/local/bin/fasthttp

# Create directories
echo -e "${GREEN}Creating directories...${NC}"
mkdir -p /fast-http
mkdir -p /var/log/fasthttp
mkdir -p /var/run

# Install configuration file if it doesn't exist
if [ ! -f "/fast-http/fasthttp.json" ]; then
    echo -e "${GREEN}Installing default configuration...${NC}"
    cp fasthttp.json /fast-http/fasthttp.json
    echo -e "${YELLOW}Configuration file installed to /fast-http/fasthttp.json${NC}"
    echo -e "${YELLOW}Please edit this file to configure your virtual hosts${NC}"
else
    echo -e "${YELLOW}Configuration file already exists at /fast-http/fasthttp.json${NC}"
fi

# Set permissions
chmod 644 /fast-http/fasthttp.json

echo -e "${GREEN}Installation complete!${NC}"
echo -e "${YELLOW}Usage:${NC}"
echo -e "  Start server:   ${GREEN}fasthttp start${NC}"
echo -e "  Stop server:    ${GREEN}fasthttp stop${NC}"
echo -e "  Check status:   ${GREEN}fasthttp status${NC}"
echo -e "  Config file:    ${GREEN}/fast-http/fasthttp.json${NC}"

