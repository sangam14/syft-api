<template>
  <div id="app">
    <header>
      <h1><span class="icon">📊</span> SBOM Generator</h1>
    </header>
    
    <div class="container">
      <div class="card generate-card">
        <h2><span class="step-number">1</span> Generate SBOM</h2>
        <div class="form-group">
          <label for="sbomSource">Source (image, directory or git URL):</label>
          <div class="input-group">
            <input 
              v-model="sbomSource" 
              type="text" 
              id="sbomSource" 
              placeholder="e.g. node:latest or ./path/to/dir"
            >
            <button 
              @click="generateSBOM" 
              :disabled="isGenerating" 
              class="primary-button"
            >
              <span v-if="isGenerating" class="loader"></span>
              {{ isGenerating ? 'Generating...' : 'Generate' }}
            </button>
          </div>
          <p class="hint">Enter a Docker image name, local directory path, or Git repository URL</p>
        </div>
        
        <div v-if="errorMessage" class="alert error-message">
          <div class="alert-header">
            <span class="alert-icon">⚠️</span>
            <h3>Error</h3>
          </div>
          <pre>{{ errorMessage }}</pre>
        </div>
        
        <div v-if="sbomResult" class="result-section">
          <div class="result-header">
            <span class="success-icon">✅</span>
            <h3>SBOM Generated Successfully</h3>
          </div>
          <div class="result-content">
            <pre>{{ sbomResult }}</pre>
          </div>
        </div>
      </div>

      <div class="card scan-card">
        <h2><span class="step-number">2</span> Scan SBOM</h2>
        <div class="options-group">
          <label class="toggle-switch">
            <input v-model="useAdvanced" type="checkbox">
            <span class="toggle-slider"></span>
            <span class="toggle-label">Use Advanced Analysis</span>
          </label>
          <button 
            @click="scanSBOM" 
            :disabled="isScanning || !sbomGenerated" 
            class="primary-button"
          >
            <span v-if="isScanning" class="loader"></span>
            {{ isScanning ? 'Scanning...' : 'Scan SBOM' }}
          </button>
        </div>
        
        <div v-if="!sbomGenerated && !scanError" class="empty-state">
          <div class="empty-icon">🔍</div>
          <p>Generate an SBOM first to enable scanning</p>
        </div>
        
        <div v-if="scanError" class="alert error-message">
          <div class="alert-header">
            <span class="alert-icon">⚠️</span>
            <h3>Scan Error</h3>
          </div>
          <pre>{{ scanError }}</pre>
          
          <div v-if="scanError.includes('ollama service is not available')" class="hint-box">
            <h4>Troubleshooting Steps:</h4>
            <ol>
              <li>Ensure Ollama is running: <code>ollama serve</code></li>
              <li>Install the required model: <code>ollama pull mistral</code></li>
              <li>Restart the application</li>
            </ol>
          </div>
        </div>
        
        <div v-if="remediationWarning" class="alert warning-message">
          <div class="alert-header">
            <span class="alert-icon">⚠️</span>
            <h3>Warning</h3>
          </div>
          <p>{{ remediationWarning }}</p>
          <p>A basic remediation script has been generated instead.</p>
        </div>
        
        <div v-if="qualityScore" class="quality-score-section">
          <div class="quality-header">
            <span class="quality-icon">🏆</span>
            <h3>SBOM Quality Score</h3>
          </div>
          
          <!-- Show appropriate content based on quality score state -->
          <div v-if="qualityScoreError" class="quality-error-content">
            <div class="error-icon">⚠️</div>
            <div class="error-details">
              <h4>Quality Score Unavailable</h4>
              <p>{{ qualityScoreError }}</p>
              <div class="install-instructions">
                <p>To enable SBOM quality scoring, the sbomqs tool needs to be installed:</p>
                <pre>wget -q https://github.com/interlynk-io/sbomqs/releases/download/v1.0.3/sbomqs-linux-amd64.tar.gz && \
tar -xzf sbomqs-linux-amd64.tar.gz && \
sudo mv sbomqs-linux-amd64/sbomqs /usr/local/bin/ && \
sudo chmod +x /usr/local/bin/sbomqs</pre>
                <a 
                  href="https://github.com/interlynk-io/sbomqs" 
                  target="_blank" 
                  class="link-button"
                >
                  Learn More About sbomqs
                </a>
              </div>
            </div>
          </div>
          
          <div v-else class="quality-content">
            <div class="score-circle" :style="scoreCircleStyle">
              <span class="score-value">{{ formattedScore }}</span>
              <span class="score-max">/10</span>
            </div>
            <div class="score-details">
              <p class="score-label">
                <span v-if="scoreValue >= 8">Excellent</span>
                <span v-else-if="scoreValue >= 6">Good</span>
                <span v-else-if="scoreValue >= 4">Average</span>
                <span v-else>Needs Improvement</span>
              </p>
              <p class="score-description">This score measures SBOM completeness, accuracy, and compliance with standards.</p>
              
              <div v-if="hasCategories" class="categories-overview">
                <h4>Category Scores:</h4>
                <ul>
                  <li v-for="(category, index) in extractCategories(qualityScore)" :key="index">
                    {{ category.name }}: <span class="category-score" :class="getCategoryScoreClass(category.score)">{{ category.score.toFixed(1) }}/10</span>
                  </li>
                </ul>
              </div>
              
              <div class="improvement-suggestions">
                <h4>Improvement Suggestions:</h4>
                <ul v-if="hasImprovementSuggestions">
                  <li v-for="(suggestion, index) in improvementSuggestions" :key="index">
                    {{ suggestion }}
                  </li>
                </ul>
                <p v-else class="no-suggestions">
                  Generate detailed improvement suggestions by clicking "Show Details"
                </p>
              </div>
              
              <div class="score-buttons">
                <button @click="toggleScoreDetails" class="secondary-button">
                  {{ showScoreDetails ? 'Hide Details' : 'Show Details' }}
                </button>
                <a 
                  href="https://github.com/interlynk-io/sbomqs" 
                  target="_blank" 
                  class="link-button"
                >
                  Learn About SBOM Quality
                </a>
              </div>
            </div>
          </div>
          
          <div v-if="showScoreDetails" class="score-details-expanded">
            <div class="details-header">
              <h4>Detailed Quality Analysis</h4>
              <p class="details-description">
                Analysis powered by <a href="https://github.com/interlynk-io/sbomqs" target="_blank">sbomqs v1.0.3</a>
              </p>
            </div>
            <pre>{{ JSON.stringify(qualityScore, null, 2) }}</pre>
          </div>
        </div>
        
        <div v-if="scanResult" class="result-section">
          <div class="result-header">
            <span class="alert-icon">🔒</span>
            <h3>Scan Results</h3>
          </div>
          <div class="result-content">
            <pre>{{ scanResult }}</pre>
          </div>
          
          <div v-if="remediationScript" class="remediation-section">
            <div class="result-header">
              <span class="success-icon">🛠️</span>
              <h3>Remediation Script</h3>
            </div>
            <div class="result-content">
              <pre>{{ remediationScript }}</pre>
              <button @click="copyToClipboard(remediationScript)" class="copy-button">
                Copy to Clipboard
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <footer>
      <p>SBOM Generator &copy; 2023</p>
    </footer>
  </div>
</template>

<script setup>
import { ref, onErrorCaptured, computed } from 'vue'

const sbomSource = ref('')
const isGenerating = ref(false)
const isScanning = ref(false)
const sbomGenerated = ref(false)
const sbomResult = ref(null)
const scanResult = ref(null)
const remediationScript = ref(null)
const remediationWarning = ref(null)
const qualityScore = ref(null)
const useAdvanced = ref(false)
const errorMessage = ref(null)
const scanError = ref(null)
const showScoreDetails = ref(false)

// Computed properties for quality score
const qualityScoreError = computed(() => {
  if (!qualityScore.value) return null
  
  // Check for specific error messages
  if (qualityScore.value.error) {
    if (qualityScore.value.error.includes('sbomqs tool not installed')) {
      return 'The sbomqs tool is not installed on the server'
    }
    return qualityScore.value.error
  }
  
  return null
})

const scoreValue = computed(() => {
  if (!qualityScore.value || qualityScoreError.value) return 0
  
  // Try to extract score from different possible formats
  if (qualityScore.value.score) {
    return parseFloat(qualityScore.value.score)
  } else if (qualityScore.value.avg_score) {
    return parseFloat(qualityScore.value.avg_score)
  } else if (qualityScore.value.files && qualityScore.value.files.length > 0) {
    return parseFloat(qualityScore.value.files[0].avg_score || 0)
  }
  
  return 0
})

const formattedScore = computed(() => {
  return scoreValue.value.toFixed(1)
})

const scoreCircleStyle = computed(() => {
  // Calculate color based on score (red to green gradient)
  const value = scoreValue.value
  let color
  
  if (value >= 8) {
    color = '#10b981' // Green for excellent
  } else if (value >= 6) {
    color = '#f59e0b' // Amber for good
  } else if (value >= 4) {
    color = '#f97316' // Orange for average
  } else {
    color = '#ef4444' // Red for poor
  }
  
  // Calculate percentage for circle fill
  const percentage = (value / 10) * 100
  
  return {
    background: `conic-gradient(${color} ${percentage}%, #e5e7eb ${percentage}% 100%)`
  }
})

// Additional computed properties for quality score display
const hasCategories = computed(() => {
  return extractCategories(qualityScore.value).length > 0
})

const hasImprovementSuggestions = computed(() => {
  return improvementSuggestions.value.length > 0
})

const improvementSuggestions = computed(() => {
  if (!qualityScore.value) return []
  
  const suggestions = []
  
  // Extract low scoring categories for improvement
  const categories = extractCategories(qualityScore.value)
  for (const category of categories) {
    if (category.score < 5) {
      suggestions.push(`Improve ${category.name} elements (${category.score.toFixed(1)}/10)`)
    }
  }
  
  // Add specific suggestions based on known patterns
  if (qualityScore.value.files && qualityScore.value.files.length > 0) {
    const file = qualityScore.value.files[0]
    
    if (file.scores) {
      for (const score of file.scores) {
        if (score.score === 0 && score.max_score > 0) {
          suggestions.push(`Add ${score.feature}: ${score.description}`)
        }
      }
    }
  }
  
  // Standard improvement suggestions if we have a low score
  if (scoreValue.value < 6) {
    suggestions.push('Ensure SBOM has supplier names for all components')
    suggestions.push('Include component relationships in SBOM')
    suggestions.push('Add license information for all components')
  }
  
  return suggestions.slice(0, 5) // Limit to top 5 suggestions
})

// Function to get class name for category score coloring
function getCategoryScoreClass(score) {
  if (score >= 8) return 'score-excellent'
  if (score >= 6) return 'score-good'
  if (score >= 4) return 'score-average'
  return 'score-poor'
}

// Add global component error handler
onErrorCaptured((err, instance, info) => {
  console.error('Component error captured:', err)
  console.error('Component:', instance)
  console.error('Error info:', info)
  
  // Set error message based on error type without referencing any external variables
  const errMsg = err?.message || 'An unexpected error occurred'
  
  // Use local reactive refs for error handling
  errorMessage.value = errMsg
  
  // Prevent error from propagating further
  return false
})

function extractCategories(scoreData) {
  if (!scoreData || !scoreData.files || !scoreData.files[0] || !scoreData.files[0].scores) {
    return []
  }
  
  const scores = scoreData.files[0].scores
  const categories = {}
  
  // Group scores by category and calculate averages
  scores.forEach(item => {
    if (!categories[item.category]) {
      categories[item.category] = {
        total: 0,
        count: 0
      }
    }
    
    categories[item.category].total += (item.score / item.max_score) * 10
    categories[item.category].count++
  })
  
  // Convert to array and calculate averages
  return Object.keys(categories).map(key => ({
    name: key,
    score: categories[key].total / categories[key].count
  }))
}

function toggleScoreDetails() {
  showScoreDetails.value = !showScoreDetails.value
}

async function generateSBOM() {
  try {
    isGenerating.value = true
    errorMessage.value = null
    sbomResult.value = null
    sbomGenerated.value = false
    qualityScore.value = null

    if (!sbomSource.value?.trim()) {
      errorMessage.value = 'Please provide a valid source'
      return
    }

    const response = await fetch('http://localhost:3000/generate-sbom', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        sbomSource: sbomSource.value
      })
    })
    
    if (!response.ok) {
      const errorText = await response.text()
      errorMessage.value = errorText || `HTTP error! status: ${response.status}`
      return
    }
    
    const data = await response.json()
    if (!data?.sbomData) {
      errorMessage.value = 'No SBOM data received from server'
      return
    }
    
    sbomResult.value = data.sbomData
    sbomGenerated.value = true
  } catch (error) {
    console.error('Error generating SBOM:', error)
    errorMessage.value = error?.message || 'An unexpected error occurred'
  } finally {
    isGenerating.value = false
  }
}

async function scanSBOM() {
  try {
    isScanning.value = true
    scanError.value = null
    scanResult.value = null
    remediationScript.value = null
    remediationWarning.value = null
    qualityScore.value = null
    showScoreDetails.value = false

    const response = await fetch('http://localhost:3000/scan-sbom', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        useAdvanced: useAdvanced.value
      })
    })
    
    if (!response.ok) {
      const errorText = await response.text()
      scanError.value = errorText || `HTTP error! status: ${response.status}`
      return
    }
    
    const data = await response.json()
    if (!data) {
      scanError.value = 'No scan data received from server'
      return
    }

    scanResult.value = data.scanResult || 'No vulnerabilities found'
    remediationScript.value = data.remediationScript || ''
    remediationWarning.value = data.remediationWarning || ''
    qualityScore.value = data.qualityScore || null
  } catch (error) {
    console.error('Error scanning SBOM:', error)
    scanError.value = error?.message || 'An unexpected error occurred'
  } finally {
    isScanning.value = false
  }
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text)
    .then(() => {
      const copyButton = document.querySelector('.copy-button')
      copyButton.textContent = '✅ Copied!'
      setTimeout(() => {
        copyButton.textContent = 'Copy to Clipboard'
      }, 2000)
    })
    .catch(err => {
      console.error('Failed to copy text: ', err)
    })
}
</script>

<style>
:root {
  --primary-color: #2563eb;
  --primary-dark: #1d4ed8;
  --success-color: #10b981;
  --warning-color: #f59e0b;
  --error-color: #ef4444;
  --text-color: #1f2937;
  --secondary-text: #4b5563;
  --bg-color: #f3f4f6;
  --card-bg: white;
  --border-color: #e5e7eb;
  --shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  background-color: var(--bg-color);
  font-family: 'Inter', 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: var(--text-color);
  line-height: 1.6;
}

#app {
  max-width: 1280px;
  margin: 0 auto;
  padding: 2rem;
}

header {
  margin-bottom: 2rem;
  text-align: center;
}

h1 {
  font-size: 2.5rem;
  color: var(--primary-color);
  margin-bottom: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon {
  margin-right: 0.5rem;
}

.container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
  gap: 2rem;
  margin-bottom: 2rem;
}

.card {
  background-color: var(--card-bg);
  border-radius: 0.75rem;
  box-shadow: var(--shadow);
  padding: 1.5rem;
  transition: transform 0.2s, box-shadow 0.2s;
}

.card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

h2 {
  font-size: 1.5rem;
  margin-bottom: 1.5rem;
  color: var(--primary-color);
  display: flex;
  align-items: center;
}

.step-number {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background-color: var(--primary-color);
  color: white;
  border-radius: 50%;
  font-size: 1rem;
  margin-right: 0.75rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.input-group {
  display: flex;
  gap: 0.5rem;
}

input[type="text"] {
  flex: 1;
  padding: 0.75rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

input[type="text"]:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.2);
}

.hint {
  font-size: 0.875rem;
  color: var(--secondary-text);
  margin-top: 0.5rem;
}

.primary-button {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 0.5rem;
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
  min-width: 120px;
}

.primary-button:hover:not(:disabled) {
  background-color: var(--primary-dark);
}

.primary-button:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.7;
}

.secondary-button {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  color: var(--primary-color);
  background-color: white;
  border: 1px solid var(--primary-color);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
}

.secondary-button:hover {
  background-color: rgba(37, 99, 235, 0.1);
}

.options-group {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

/* Toggle switch styling */
.toggle-switch {
  position: relative;
  display: inline-flex;
  align-items: center;
  cursor: pointer;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 26px;
  background-color: #ccc;
  border-radius: 34px;
  transition: .4s;
  margin-right: 10px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  border-radius: 50%;
  transition: .4s;
}

input:checked + .toggle-slider {
  background-color: var(--primary-color);
}

input:checked + .toggle-slider:before {
  transform: translateX(24px);
}

.toggle-label {
  font-weight: 500;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--secondary-text);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.alert {
  margin: 1.5rem 0;
  border-radius: 0.5rem;
  padding: 1rem;
}

.error-message {
  background-color: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.warning-message {
  background-color: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.alert-header {
  display: flex;
  align-items: center;
  margin-bottom: 0.5rem;
}

.alert-icon, .success-icon, .quality-icon {
  margin-right: 0.5rem;
  font-size: 1.25rem;
}

.result-section, .quality-score-section {
  margin-top: 1.5rem;
}

.result-header, .quality-header {
  display: flex;
  align-items: center;
  margin-bottom: 0.75rem;
}

.result-content {
  position: relative;
}

pre {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1rem;
  overflow-x: auto;
  font-family: 'Fira Mono', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 400px;
  overflow-y: auto;
}

.remediation-section {
  margin-top: 1.5rem;
}

.copy-button {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  background-color: white;
  border: 1px solid var(--border-color);
  border-radius: 0.25rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.copy-button:hover {
  background-color: #f3f4f6;
}

.hint-box {
  margin-top: 1rem;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1rem;
}

.hint-box h4 {
  margin-bottom: 0.5rem;
  color: var(--secondary-text);
}

.hint-box ol {
  padding-left: 1.5rem;
}

.hint-box li {
  margin-bottom: 0.5rem;
}

.hint-box code {
  background-color: #e5e7eb;
  padding: 0.15rem 0.3rem;
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
}

/* Quality Score Styling */
.quality-content {
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.5rem;
}

.score-circle {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: conic-gradient(#10b981 70%, #e5e7eb 70% 100%);
  position: relative;
  flex-shrink: 0;
}

.score-circle::before {
  content: "";
  position: absolute;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background-color: white;
}

.score-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--text-color);
  z-index: 1;
}

.score-max {
  font-size: 0.875rem;
  color: var(--secondary-text);
  z-index: 1;
  margin-top: -0.25rem;
}

.score-details {
  flex: 1;
}

.score-label {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.score-description {
  color: var(--secondary-text);
  margin-bottom: 1rem;
}

.categories-overview {
  margin-top: 1rem;
  margin-bottom: 1rem;
}

.categories-overview h4 {
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: var(--secondary-text);
}

.categories-overview ul {
  list-style: none;
}

.categories-overview li {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
}

.category-score {
  font-weight: 600;
}

.score-details-expanded {
  margin-top: 1rem;
}

/* Loading spinner */
.loader {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s linear infinite;
  margin-right: 0.5rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

footer {
  text-align: center;
  margin-top: 3rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
  color: var(--secondary-text);
  font-size: 0.875rem;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .container {
    grid-template-columns: 1fr;
  }
  
  .input-group {
    flex-direction: column;
  }
  
  .options-group {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .primary-button {
    width: 100%;
  }
  
  .quality-content {
    flex-direction: column;
    align-items: center;
  }
  
  .score-details {
    text-align: center;
  }
}

/* New styles for improved quality score display */
.score-buttons {
  display: flex;
  gap: 0.75rem;
  margin-top: 1rem;
}

.link-button {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  color: var(--secondary-text);
  background-color: white;
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  transition: all 0.2s;
}

.link-button:hover {
  background-color: #f3f4f6;
  color: var(--primary-color);
}

.details-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.details-description {
  font-size: 0.875rem;
  color: var(--secondary-text);
}

.details-description a {
  color: var(--primary-color);
  text-decoration: none;
}

.improvement-suggestions {
  margin-top: 1rem;
}

.improvement-suggestions h4 {
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: var(--secondary-text);
}

.improvement-suggestions ul {
  list-style: circle;
  padding-left: 1.25rem;
}

.improvement-suggestions li {
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
}

.no-suggestions {
  font-size: 0.875rem;
  color: var(--secondary-text);
  font-style: italic;
}

/* Category score colors */
.score-excellent {
  color: var(--success-color);
}

.score-good {
  color: var(--warning-color);
}

.score-average {
  color: #f97316; /* Orange */
}

.score-poor {
  color: var(--error-color);
}

/* Add styles for quality score error */
.quality-error-content {
  display: flex;
  gap: 1.5rem;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.5rem;
}

.error-icon {
  font-size: 2rem;
  color: var(--warning-color);
}

.error-details {
  flex: 1;
}

.error-details h4 {
  font-size: 1.1rem;
  margin-bottom: 0.5rem;
  color: var(--warning-color);
}

.error-details p {
  margin-bottom: 1rem;
}

.install-instructions {
  margin-top: 1rem;
  padding: 1rem;
  background-color: #f1f5f9;
  border-radius: 0.5rem;
}

.install-instructions pre {
  background-color: #e2e8f0;
  padding: 0.75rem;
  margin: 0.75rem 0;
  border-radius: 0.375rem;
  font-size: 0.8rem;
  overflow-x: auto;
}
</style>