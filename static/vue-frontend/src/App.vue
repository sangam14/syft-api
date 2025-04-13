<template>
  <div id="app">
    <!-- Navigation Sidebar -->
    <nav class="sidebar">
      <div class="sidebar-header">
        <span class="icon">📊</span>
        <h1>SBOM Generator</h1>
      </div>
      <div class="nav-links">
        <a href="#generate" class="nav-link active">
          <span class="nav-icon">🔧</span>
          Generate SBOM
        </a>
        <a href="#analytics" class="nav-link">
          <span class="nav-icon">📈</span>
          Analytics
        </a>
        <a href="#scan" class="nav-link">
          <span class="nav-icon">🔍</span>
          Scan SBOM
        </a>
        <a href="#logs" class="nav-link">
          <span class="nav-icon">📝</span>
          View Logs
        </a>
      </div>
      <div class="sidebar-footer">
        <p>Version 1.0.0</p>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="main-content">
      <div class="container">
        <!-- Generate SBOM Card -->
        <div class="card generate-card" id="generate">
          <div class="card-header">
            <span class="step-number">1</span>
            <h2>Generate SBOM</h2>
          </div>
          <div class="card-body">
            <div class="form-group">
              <label for="sbomSource">Source (image, directory or git URL):</label>
              <div class="input-group">
                <input
                  v-model="sbomSource"
                  type="text"
                  id="sbomSource"
                  placeholder="e.g. node:latest or ./path/to/dir"
                  class="modern-input"
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
        </div>

        <!-- Analytics Section -->
        <div class="card analytics-card" id="analytics">
          <div class="card-header">
            <span class="step-number">2</span>
            <h2>Analytics</h2>
          </div>
          <div class="card-body">
            <div v-if="!sbomGenerated" class="empty-state">
              <div class="empty-icon">📊</div>
              <p>Generate an SBOM first to view analytics</p>
            </div>

            <div v-else>
              <!-- Full Analytics Dashboard -->

              <analytics-view
                :sbom-data="sbomData"
                @navigate="handleNavigation"
              />
            </div>
          </div>
        </div>

        <!-- Scan SBOM Card -->
        <div class="card scan-card" id="scan">
          <div class="card-header">
            <span class="step-number">3</span>
            <h2>Scan SBOM</h2>
          </div>
          <div class="card-body">
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
                  <button
                    @click="copyToClipboard(remediationScript, $event)"
                    class="copy-button">
                    Copy to Clipboard
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onErrorCaptured, computed, defineAsyncComponent } from 'vue'

// Lazy load components for better performance
// These components are now used inside the AnalyticsView component
// We don't need to import them directly in App.vue anymore
const AnalyticsView = defineAsyncComponent(() => import('./components/AnalyticsView.vue'))

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

// Navigation handler for components
function handleNavigation(target) {
  // Scroll to the target section
  const targetSection = document.getElementById(target);
  if (targetSection) {
    targetSection.scrollIntoView({ behavior: 'smooth' });

    // Update active nav link
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => {
      link.classList.remove('active');
      if (link.getAttribute('href') === `#${target}`) {
        link.classList.add('active');
      }
    });
  }
}

// API service with caching and retry logic
const apiCache = new Map()
const API_BASE_URL = 'http://localhost:3000'

async function fetchWithRetry(url, options, retries = 3, delay = 500) {
  try {
    const response = await fetch(url, options)
    if (response.ok) return response

    if (retries > 0 && [408, 429, 500, 502, 503, 504].includes(response.status)) {
      await new Promise(resolve => setTimeout(resolve, delay))
      return fetchWithRetry(url, options, retries - 1, delay * 2)
    }

    return response
  } catch (error) {
    if (retries > 0) {
      await new Promise(resolve => setTimeout(resolve, delay))
      return fetchWithRetry(url, options, retries - 1, delay * 2)
    }
    throw error
  }
}

async function apiRequest(endpoint, method = 'GET', data = null, useCache = false) {
  const url = `${API_BASE_URL}/${endpoint}`
  const cacheKey = `${method}:${url}:${JSON.stringify(data)}`

  // Return cached response if available and cache is enabled
  if (useCache && apiCache.has(cacheKey)) {
    return apiCache.get(cacheKey)
  }

  const options = {
    method,
    headers: {
      'Content-Type': 'application/json'
    },
    body: data ? JSON.stringify(data) : undefined
  }

  const response = await fetchWithRetry(url, options)

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || `HTTP error! status: ${response.status}`)
  }

  const result = await response.json()

  // Cache the response if cache is enabled
  if (useCache) {
    apiCache.set(cacheKey, result)
  }

  return result
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

    const data = await apiRequest('generate-sbom', 'POST', {
      sbomSource: sbomSource.value
    })

    if (!data?.sbomData) {
      errorMessage.value = 'No SBOM data received from server'
      return
    }

    // Store the raw SBOM data
    sbomResult.value = data.sbomData

    // Set the generated flag to true
    sbomGenerated.value = true

    // Log success for debugging
    console.log('SBOM generated successfully, data available:', !!sbomData.value)
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

function copyToClipboard(text, event) {
  try {
    if (!text) return;

    // Make a local copy of any variables needed to prevent reference errors
    const textToCopy = String(text);
    const currentEvent = event || null;

    // Use a single function to update button text for better code reuse
    const updateButtonText = (button, success = true) => {
      if (!button) return;
      const originalText = button.textContent || 'Copy to Clipboard';
      button.textContent = success ? '✅ Copied!' : '❌ Failed';
      setTimeout(() => {
        button.textContent = originalText;
      }, 2000);
    };

    // Try the modern Clipboard API first
    navigator.clipboard.writeText(textToCopy)
      .then(() => {
        if (currentEvent?.target) {
          updateButtonText(currentEvent.target, true);
        }
      })
      .catch(err => {
        console.error('Clipboard API failed:', err);

        // Fallback for browsers without clipboard API
        const textarea = document.createElement('textarea');
        textarea.value = textToCopy;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.focus();
        textarea.select();

        try {
          // Use execCommand as fallback (deprecated but still works in most browsers)
          const success = document.execCommand('copy');

          if (currentEvent?.target) {
            updateButtonText(currentEvent.target, success);
          }
        } catch (execError) {
          console.error('Fallback copy failed:', execError);
          if (currentEvent?.target) {
            updateButtonText(currentEvent.target, false);
          }
        } finally {
          document.body.removeChild(textarea);
        }
      });
  } catch (error) {
    console.error('Error in copyToClipboard function:', error);
  }
}

const sbomData = computed(() => {
  if (!sbomResult.value) return null;
  try {
    // If sbomResult is a string (JSON), parse it
    if (typeof sbomResult.value === 'string') {
      return JSON.parse(sbomResult.value);
    }

    // If it's already an object, return it
    if (typeof sbomResult.value === 'object') {
      return sbomResult.value;
    }

    return null;
  } catch (error) {
    console.error('Error parsing SBOM data:', error);
    return null;
  }
});
</script>

<style>
:root {
  /* Modern Color Palette */
  --primary-color: #0d9488; /* Teal/Cyan */
  --primary-hover: #0f766e; /* Darker Teal/Cyan */
  --secondary-color: #475569;
  --success-color: #059669;
  --danger-color: #dc2626;
  --warning-color: #d97706;
  --info-color: #0d9488; /* Match primary */
  --light-color: #f8fafc;
  --dark-color: #111827; /* Slightly darker */
  --border-color: #e2e8f0;
  --sidebar-width: 260px; /* Slightly narrower */
  --header-height: 70px;
  --card-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  --transition-speed: 0.3s;

  /* Enhanced Typography */
  --font-sans: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif; /* Modern system font stack */
  --font-mono: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
  --font-size-xs: 0.75rem;
  --font-size-sm: 0.875rem;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;
  --font-size-xl: 1.25rem;
  --font-size-2xl: 1.5rem;
  --font-size-3xl: 1.875rem;
  --font-size-4xl: 2.25rem;
  --font-weight-normal: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  --font-weight-bold: 700;
  --line-height-normal: 1.6;
  --line-height-tight: 1.3;
  --letter-spacing-tight: -0.025em;
  --letter-spacing-normal: 0;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: var(--font-sans);
  font-size: var(--font-size-base);
  line-height: var(--line-height-normal);
  color: var(--dark-color);
  background-color: #f8fafc;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  display: flex;
  min-height: 100vh;
}

/* Enhanced Sidebar Styles */
.sidebar {
  width: var(--sidebar-width);
  background: linear-gradient(to bottom, var(--dark-color), #1e293b);
  color: white;
  padding: 2rem;
  display: flex;
  flex-direction: column;
  position: fixed;
  height: 100vh;
  left: 0;
  top: 0;
  box-shadow: 4px 0 15px rgba(0, 0, 0, 0.15);
  z-index: 100;
}

.sidebar-header {
  padding: 1.5rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.15);
  margin-bottom: 2.5rem;
  position: relative;
}

.sidebar-header h1 {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  margin-top: 0.75rem;
  letter-spacing: var(--letter-spacing-tight);
  line-height: var(--line-height-tight);
  background: linear-gradient(135deg, #ffffff, #cbd5e1);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  position: relative;
}

.sidebar-header .icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: 0.5rem;
  filter: drop-shadow(0 4px 6px rgba(0, 0, 0, 0.5));
}

.nav-links {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.nav-link {
  display: flex;
  align-items: center;
  padding: 1rem 1.25rem;
  color: rgba(255, 255, 255, 0.85);
  text-decoration: none;
  border-radius: 12px;
  margin-bottom: 0.75rem;
  transition: all 0.3s;
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-lg);
  position: relative;
  overflow: hidden;
}

.nav-link:after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 1.25rem;
  width: 0;
  height: 2px;
  background-color: white;
  transition: width 0.3s ease;
}

.nav-link:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: white;
  transform: translateX(4px);
}

.nav-link:hover:after {
  width: calc(100% - 2.5rem);
}

.nav-link.active {
  background: linear-gradient(135deg, var(--primary-color), #3b82f6);
  color: white;
  box-shadow: 0 10px 15px -3px rgba(37, 99, 235, 0.3);
}

.nav-link.active:after {
  width: 0;
}

.nav-icon {
  margin-right: 1rem;
  font-size: 1.5rem;
  transition: transform 0.3s ease;
}

.nav-link:hover .nav-icon {
  transform: translateY(-2px);
}

.sidebar-footer {
  padding: 1.5rem 0;
  border-top: 1px solid rgba(255, 255, 255, 0.15);
  font-size: var(--font-size-sm);
  color: rgba(255, 255, 255, 0.6);
  text-align: center;
}

/* Enhanced Main Content Styles */
.main-content {
  flex: 1;
  margin-left: var(--sidebar-width);
  padding: 3rem;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 1fr; /* Default to single column */
  gap: 2rem; /* Add gap between grid items */
}

/* Enhanced Card Styles */
.card {
  background: white;
  border-radius: 14px;
  box-shadow: 0 8px 15px -3px rgba(0, 0, 0, 0.08), 0 4px 6px -2px rgba(0, 0, 0, 0.04);
  margin-bottom: 2rem;
  overflow: hidden;
  border: 1px solid var(--border-color);
  transition: all 0.3s ease;
  position: relative;
  /* Remove max-width and centering margins from individual cards */
  /* Let the grid container manage the layout */
}

.card:hover {
  transform: translateY(-4px);
  box-shadow: 0 15px 20px -5px rgba(0, 0, 0, 0.08), 0 10px 10px -5px rgba(0, 0, 0, 0.03);
}

.card-header {
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  background: linear-gradient(to right, var(--light-color), white);
}

.card-header h2 {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--dark-color);
  letter-spacing: var(--letter-spacing-tight);
  position: relative;
}

.card-header h2::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: -8px;
  width: 40px;
  height: 3px;
  background-color: var(--primary-color);
  border-radius: 2px;
}

.step-number {
  background: linear-gradient(135deg, var(--primary-color), var(--primary-hover));
  color: white;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 1rem;
  font-weight: var(--font-weight-bold);
  font-size: var(--font-size-lg);
  box-shadow: 0 6px 12px -4px rgba(37, 99, 235, 0.3);
}

.card-body {
  padding: 2rem;
  position: relative;
}

/* Enhanced Form Styles */
.form-group {
  margin-bottom: 2rem;
}

label {
  display: block;
  margin-bottom: 0.75rem;
  font-weight: var(--font-weight-medium);
  color: var(--dark-color);
  font-size: var(--font-size-lg);
  position: relative;
  transition: all 0.2s;
}

.modern-input {
  width: 100%;
  padding: 1.25rem 1.5rem;
  border: 2px solid var(--border-color);
  border-radius: 12px;
  font-size: var(--font-size-base);
  transition: all 0.3s;
  font-family: var(--font-sans);
  line-height: var(--line-height-normal);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  background-color: white;
}

.modern-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.15);
  transform: translateY(-2px);
}

.modern-input::placeholder {
  color: var(--secondary-color);
  opacity: 0.7;
}

.input-group {
  display: flex;
  gap: 1.25rem;
  position: relative;
}

.input-group input {
  flex: 1;
}

/* Enhanced Button Styles */
.primary-button {
  background: linear-gradient(135deg, var(--primary-color), var(--primary-hover));
  color: white;
  border: none;
  padding: 1.25rem 2rem;
  border-radius: 12px;
  cursor: pointer;
  font-weight: var(--font-weight-semibold);
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 160px;
  font-size: var(--font-size-lg);
  box-shadow: 0 10px 15px -3px rgba(37, 99, 235, 0.3);
  position: relative;
  overflow: hidden;
}

.primary-button:before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(to right, transparent 0%, rgba(255, 255, 255, 0.2) 50%, transparent 100%);
  transform: translateX(-100%);
}

.primary-button:hover {
  background: linear-gradient(135deg, var(--primary-hover), var(--primary-color));
  transform: translateY(-5px);
  box-shadow: 0 20px 25px -5px rgba(37, 99, 235, 0.4);
}

.primary-button:hover:before {
  transform: translateX(100%);
  transition: transform 0.8s ease;
}

.primary-button:active {
  transform: translateY(-2px);
  box-shadow: 0 5px 10px -3px rgba(37, 99, 235, 0.5);
}

.primary-button:disabled {
  background: linear-gradient(135deg, var(--secondary-color), #64748b);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
  opacity: 0.7;
}

.secondary-button {
  background-color: white;
  color: var(--primary-color);
  border: 2px solid var(--primary-color);
  padding: 1rem 1.75rem;
  border-radius: 12px;
  cursor: pointer;
  font-weight: var(--font-weight-semibold);
  transition: all 0.3s;
  font-size: var(--font-size-lg);
  position: relative;
  overflow: hidden;
  z-index: 1;
}

.secondary-button:before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: var(--primary-color);
  z-index: -1;
  transform: scaleX(0);
  transform-origin: right;
  transition: transform 0.3s ease;
}

.secondary-button:hover {
  color: white;
  transform: translateY(-3px);
  box-shadow: 0 10px 15px -3px rgba(37, 99, 235, 0.2);
}

.secondary-button:hover:before {
  transform: scaleX(1);
  transform-origin: left;
}

/* Enhanced Alert Styles */
.alert {
  padding: 1.5rem;
  border-radius: 12px;
  margin-bottom: 1.5rem;
  font-size: var(--font-size-base);
  position: relative;
  animation: slideIn 0.4s ease-out;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
}

.error-message {
  background-color: rgba(220, 38, 38, 0.05);
  border-left: 4px solid var(--danger-color);
}

.warning-message {
  background-color: rgba(217, 119, 6, 0.05);
  border-left: 4px solid var(--warning-color);
}

.alert-header {
  display: flex;
  align-items: center;
  margin-bottom: 0.75rem;
}

.alert-icon {
  margin-right: 0.75rem;
  font-size: 1.5rem;
  animation: pulseAlert 2s infinite;
}

/* Enhanced Result Section Styles */
.result-section {
  margin-top: 2rem;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
  animation: fadeIn 0.5s ease-out;
}

.result-section:hover {
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  transform: translateY(-3px);
}

.result-header {
  background: linear-gradient(to right, var(--light-color), white);
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
}

.result-header h3 {
  margin: 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  margin-left: 1rem;
  position: relative;
}

.result-content {
  padding: 1.5rem;
  position: relative;
}

.result-content pre {
  background-color: var(--light-color);
  padding: 1.5rem;
  border-radius: 12px;
  overflow-x: auto;
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-tight);
  border: 1px solid var(--border-color);
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.05);
}

/* Enhanced Quality Score Styles */
.quality-score-section {
  margin-top: 2rem;
}

.score-circle {
  width: 160px;
  height: 160px;
  border-radius: 50%;
  background-color: var(--primary-color);
  color: white;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin: 0 auto 2rem;
  box-shadow: 0 8px 12px -1px rgba(0, 0, 0, 0.1);
}

.score-value {
  font-size: 3.5rem;
  font-weight: var(--font-weight-bold);
  line-height: 1;
}

.score-max {
  font-size: var(--font-size-xl);
  opacity: 0.8;
  margin-top: 0.5rem;
}

.score-details {
  text-align: center;
}

.score-label {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  margin-bottom: 0.75rem;
}

.score-description {
  color: var(--secondary-color);
  margin-bottom: 1.5rem;
  font-size: var(--font-size-lg);
  line-height: var(--line-height-normal);
}

/* Enhanced Toggle Switch Styles */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 60px;
  height: 28px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--secondary-color);
  transition: .4s;
  border-radius: 28px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .toggle-slider {
  background-color: var(--primary-color);
}

input:checked + .toggle-slider:before {
  transform: translateX(32px);
}

.toggle-label {
  margin-left: 1.25rem;
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-lg);
}

/* Enhanced Empty State Styles */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: var(--secondary-color);
  background: linear-gradient(to bottom, white, var(--light-color));
  border-radius: 16px;
  border: 1px dashed var(--border-color);
  transition: all 0.3s ease;
  animation: pulse 2s infinite ease-in-out;
}

.empty-state:hover {
  border-color: var(--primary-color);
}

.empty-icon {
  font-size: 5rem;
  margin-bottom: 1.5rem;
  opacity: 0.7;
}

.empty-state p {
  font-size: var(--font-size-lg);
  max-width: 400px;
  margin: 0 auto;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(37, 99, 235, 0.1);
  }
  70% {
    box-shadow: 0 0 0 15px rgba(37, 99, 235, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(37, 99, 235, 0);
  }
}

/* Enhanced Responsive Design */
@media (min-width: 1024px) { /* Apply 2-column layout on larger screens */
  .container {
    grid-template-columns: repeat(2, 1fr);
  }

  /* Make the generate card span both columns if desired, or adjust as needed */
  .generate-card {
     grid-column: span 2; /* Example: Make generate card wider */
  }

  /* Adjust analytics grid items if the container is now 2 columns */
   .analytics-grid {
     grid-template-columns: repeat(2, 1fr);
     gap: 1.5rem;
   }
}

@media (max-width: 1280px) {
  .analytics-grid {
    gap: 1.5rem;
  }
  /* Adjust container gap for slightly smaller screens if needed */
  .container {
     gap: 1.5rem;
  }
}

@media (max-width: 1024px) {
  .sidebar {
    width: 240px;
    padding: 1.5rem 1rem;
  }

  .main-content {
    margin-left: 240px;
    padding: 2rem;
  }

  .analytics-grid > *:nth-child(1),
  .analytics-grid > *:nth-child(2) {
    grid-column: span 2;
  }
}

@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    height: auto;
    position: relative;
    padding: 1rem;
  }

  .sidebar-header {
    padding: 1rem 0;
    margin-bottom: 1rem;
    text-align: center;
  }

  .sidebar-header h1 {
    font-size: var(--font-size-2xl);
  }

  .nav-links {
    flex-direction: row;
    justify-content: center;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .nav-link {
    padding: 0.75rem 1rem;
    margin-bottom: 0;
    font-size: var(--font-size-base);
  }

  .nav-icon {
    margin-right: 0.5rem;
    font-size: 1.25rem;
  }

  .sidebar-footer {
    display: none;
  }

  .main-content {
    margin-left: 0;
    padding: 1.5rem;
  }

  .input-group {
    flex-direction: column;
  }

  .primary-button {
    width: 100%;
  }

  .card-header {
    padding: 1.5rem;
  }

  .card-body {
    padding: 1.5rem;
  }

  .analytics-grid {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .analytics-grid > * {
    grid-column: span 1;
  }
}

@media (max-width: 480px) {
  .container {
    padding: 0;
  }

  .card {
    border-radius: 0;
    border-left: none;
    border-right: none;
    box-shadow: 0 5px 10px -3px rgba(0, 0, 0, 0.1);
  }

  .card-header {
    padding: 1.25rem;
  }

  .card-body {
    padding: 1.25rem;
  }

  .step-number {
    width: 36px;
    height: 36px;
    font-size: var(--font-size-base);
    margin-right: 0.75rem;
  }

  .card-header h2 {
    font-size: var(--font-size-xl);
  }
}

/* Enhanced Animation */
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loader {
  display: inline-block;
  width: 24px;
  height: 24px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s ease-in-out infinite;
  margin-right: 0.75rem;
}

/* Enhanced Typography Styles */
h1, h2, h3, h4, h5, h6 {
  font-weight: var(--font-weight-semibold);
  line-height: var(--line-height-tight);
  color: var(--dark-color);
  letter-spacing: var(--letter-spacing-tight);
}

h1 { font-size: var(--font-size-4xl); }
h2 { font-size: var(--font-size-3xl); }
h3 { font-size: var(--font-size-2xl); }
h4 { font-size: var(--font-size-xl); }
h5 { font-size: var(--font-size-lg); }
h6 { font-size: var(--font-size-base); }

p {
  margin-bottom: 1.25rem;
  line-height: var(--line-height-normal);
  font-size: var(--font-size-base);
}

pre, code {
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-tight);
}

/* Enhanced Link Styles */
a {
  color: var(--primary-color);
  text-decoration: none;
  transition: all var(--transition-speed);
  font-weight: var(--font-weight-medium);
}

a:hover {
  color: var(--primary-hover);
  text-decoration: underline;
}

/* Enhanced Copy Button Styles */
.copy-button {
  position: absolute;
  top: 1rem;
  right: 1rem;
  background-color: white;
  border: 2px solid var(--border-color);
  border-radius: 8px;
  padding: 0.75rem 1.25rem;
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: all 0.3s;
  font-weight: var(--font-weight-medium);
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.copy-button:before {
  content: '📋';
  font-size: 1rem;
}

.copy-button:hover {
  background-color: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
  transform: translateY(-3px);
  box-shadow: 0 10px 15px -3px rgba(37, 99, 235, 0.3);
}

.copy-button:hover:before {
  content: '✂️';
}

/* Enhanced Animations */
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideIn {
  from { opacity: 0; transform: translateX(-20px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes pulseAlert {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

.analytics-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
  margin-top: 1.5rem;
  animation: fadeIn 0.6s ease-out;
}

.analytics-grid > * {
  grid-column: span 2;
  transition: all 0.3s ease;
  border-radius: 16px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.analytics-grid > *:hover {
  transform: translateY(-5px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.analytics-grid > *:nth-child(1),
.analytics-grid > *:nth-child(2) {
  grid-column: span 1;
}

/* Package Metrics Section Styles */
.package-metrics-section {
  margin-bottom: 2rem;
  animation: fadeIn 0.4s ease-out;
}

.section-header {
  margin-bottom: 1.5rem;
}

.section-header h3 {
  font-size: var(--font-size-2xl);
  margin-bottom: 0.5rem;
  position: relative;
  display: inline-block;
}

.section-header h3::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: -8px;
  width: 40px;
  height: 3px;
  background-color: var(--primary-color);
  border-radius: 2px;
}

.section-description {
  color: var(--secondary-color);
  font-size: var(--font-size-lg);
  max-width: 800px;
}

.section-divider {
  height: 1px;
  background: linear-gradient(to right, var(--border-color), transparent);
  margin: 2rem 0;
}

/* Enhance Package Metrics component appearance */
.package-metrics {
  border-radius: 16px;
  box-shadow: 0 10px 20px -5px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  transition: all 0.3s ease;
  background: white;
}

.package-metrics:hover {
  transform: translateY(-5px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}
</style>