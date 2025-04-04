const outputDiv = document.getElementById('output');
const logTable = document.getElementById('logTableBody');

// Utility to update output area
const updateOutput = (text, isError = false) => {
    outputDiv.textContent = text;
    outputDiv.style.color = isError ? 'red' : 'black';
};

// Fetch and render logs
const fetchLogs = async () => {
    try {
        const response = await fetch('/logs');
        const data = await response.text();
        const logLines = data.split("\n").filter(line => line.trim() !== "");
        logTable.innerHTML = "";

        logLines.forEach(line => {
            const row = document.createElement("tr");
            row.classList.add("log-row");

            const parts = line.split(" ");
            const timestamp = parts.slice(0, 2).join(" ");
            const message = parts.slice(2).join(" ");

            const timestampCell = document.createElement("td");
            timestampCell.textContent = timestamp;

            const messageCell = document.createElement("td");
            messageCell.textContent = message;

            row.appendChild(timestampCell);
            row.appendChild(messageCell);
            logTable.appendChild(row);
        });
    } catch (error) {
        console.error('Error fetching logs:', error);
    }
};

// Generate SBOM from user input
const generateSBOM = async () => {
    const source = document.getElementById('sbomInput').value;
    updateOutput('Generating SBOM...');

    try {
        const response = await fetch(`/generate-sbom?source=${encodeURIComponent(source)}`);
        const data = await response.json();
        updateOutput(JSON.stringify(data, null, 2));
    } catch (error) {
        console.error("Generate SBOM API error:", error);
        updateOutput(`Error: ${error}`, true);
    }
};

// Scan SBOM and display result
const scanSBOM = async () => {
    updateOutput('Scanning SBOM...');

    try {
        const response = await fetch('/scan-sbom');
        const data = await response.json();
        updateOutput(JSON.stringify(data, null, 2));
    } catch (error) {
        console.error("Scan SBOM API error:", error);
        updateOutput(`Error: ${error}`, true);
    }
};

// Auto-refresh logs every 2 seconds
setInterval(fetchLogs, 2000);