function simulateProgress(callback) {
    document.getElementById("progressContainer").style.display = "block";
    const bar = document.getElementById("progressBar");
    let progress = 0;
    const interval = setInterval(() => {
        progress += 10;
        bar.style.width = progress + "%";
        bar.textContent = progress + "%";
        if (progress >= 100) {
            clearInterval(interval);
            callback();
        }
    }, 100);
}

async function generateSBOM() {
    simulateProgress(async () => {
        try {
            const res = await fetch('/generate-sbom', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    sbomSource: document.getElementById("sbomInput").value
                })
            });
            if (!res.ok) {
                throw new Error(`HTTP error! status: ${res.status}`);
            }
            const data = await res.json();
            document.getElementById("scanResult").textContent = data.message || data.error || 'Unknown error';
            document.getElementById("results").style.display = "block";
        } catch (error) {
            console.error("Error generating SBOM:", error);
            document.getElementById("scanResult").textContent = "Error generating SBOM. Check console for details.";
            document.getElementById("results").style.display = "block";
        } finally {
            document.getElementById("progressContainer").style.display = "none";
            document.getElementById("progressBar").style.width = "0%";
            document.getElementById("progressBar").textContent = "0%";
        }
    });
}

async function scanSBOM() {
    document.getElementById("results").style.display = "block";
    simulateProgress(async () => {
        try {
            const res = await fetch('/scan-sbom', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    sbomSource: document.getElementById("sbomInput").value
                })
            });
            if (!res.ok) {
                throw new Error(`HTTP error! status: ${res.status}`);
            }
            const data = await res.json();

            if (!data.scanResult && !data.remediationScript && !data.remediationCommands) {
                document.getElementById("scanResult").textContent = data.message || data.error || 'Unknown error';
                return;
            }

            document.getElementById("scanResult").textContent = data.scanResult || '';
            document.getElementById("remediationScript").textContent = data.remediationScript || '';
            document.getElementById("remediationCommands").textContent = data.remediationCommands || '';

            // Render markdown scan result
            if (data.scanResultMarkdown) {
                document.getElementById("scanResultTable").innerHTML = marked.parse(data.scanResultMarkdown);
            }
            
            // Enable copy and download buttons if there are remediation commands
            const remediationCommands = document.getElementById("remediationCommands").textContent;
            document.getElementById("copyButton").disabled = !remediationCommands;
            document.getElementById("downloadButton").disabled = !remediationCommands;
        } catch (error) {
            console.error("Error scanning SBOM:", error);
            document.getElementById("scanResult").textContent = "Error scanning SBOM. Check console for details.";
        } finally {
            document.getElementById("progressContainer").style.display = "none";
            document.getElementById("progressBar").style.width = "0%";
            document.getElementById("progressBar").textContent = "0%";
        }
    });
}

function copyToClipboard() {
    const scriptText = document.getElementById("remediationCommands").innerText;
    navigator.clipboard.writeText(scriptText).then(() => {
        alert("Script copied to clipboard!");
    }).catch(err => {
        alert("Failed to copy script: " + err);
    });
}

function downloadScript() {
    const scriptText = document.getElementById("remediationCommands").innerText;
    const blob = new Blob([scriptText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "vulnerability_fix.sh";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
}
