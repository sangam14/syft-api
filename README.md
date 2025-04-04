# AI-Powered SBOM Generator

This project provides an AI-powered Software Bill of Materials (SBOM) generator.

## Current Features

*   **Ollama Integration:**  Supports generating SBOMs using the Ollama model runner.
*   **Docker Compose Support:**  Can be easily deployed and run using Docker Compose.
*   **Docker Model Runner (Planned):** Future support for running models directly within Docker.

## Prerequisites

*  ollama running locally 
*  Docker Docker and Docker Compose installed.

## Installation and Setup

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/sangam14/syft-api
    cd syft-api
    ```

2.  **Build the Docker Images:**
    ```bash
    docker compose build --no-cache
    ```

3.  **Start the Application:**
    ```bash
    docker compose up
    ```

## Usage

This tool can scan various sources to generate SBOMs:

*   **Docker Images:** Analyze existing Docker images.
*   **Directories:** Scan local directories for software components.
*   **GitHub Repositories:** Generate SBOMs from GitHub repositories.

## Accessing the API

The OpenAPI specification and interface are available at:

[http://localhost:3000](http://localhost:3000)

## Future Enhancements

*   Docker Model Runner support.
*   More model support.
*   More source support.
*   More output formats.

## Contributing

Contributions to this project are welcome! Please see the CONTRIBUTING.md file for guidelines on how to contribute.