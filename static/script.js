function fetchLogs() {
    fetch('/logs')
        .then(response => response.text())
        .then(data => {
            const logTable = document.getElementById('logTableBody');
            logTable.innerHTML = ""; // Clear existing logs

            const logLines = data.split("\n").filter(line => line.trim() !== ""); // Remove empty lines
            logLines.forEach(line => {
                const row = document.createElement("tr");
                const parts = line.split(" "); // Split log line by spaces
                const timestamp = parts.slice(0, 2).join(" "); // Extract timestamp
                const message = parts.slice(2).join(" "); // Extract log message

                const timestampCell = document.createElement("td");
                timestampCell.textContent = timestamp;

                const messageCell = document.createElement("td");
                messageCell.textContent = message;

                row.appendChild(timestampCell);
                row.appendChild(messageCell);
                logTable.appendChild(row);
            });
        })
        .catch(error => console.error('Error fetching logs:', error));
}

// Refresh logs every 2 seconds
setInterval(fetchLogs, 2000);