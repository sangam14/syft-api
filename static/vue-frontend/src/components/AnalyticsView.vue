<template>
  <div class="analytics-view">
    <div class="analytics-header">
      <div class="header-content">
        <h2>SBOM Analytics Dashboard</h2>
        <p class="description">Comprehensive analysis and insights for your Software Bill of Materials</p>
      </div>
      <div class="header-actions">
        <div class="time-range-selector">
          <label for="timeRange">Time Range:</label>
          <select id="timeRange" v-model="selectedTimeRange" class="modern-input">
            <option value="7">Last 7 Days</option>
            <option value="30">Last 30 Days</option>
            <option value="90">Last 90 Days</option>
            <option value="365">Last Year</option>
            <option value="all">All Time</option>
          </select>
        </div>
        <button @click="refreshData" class="refresh-button">
          <span class="refresh-icon">🔄</span>
          Refresh Data
        </button>
      </div>
    </div>

    <!-- Debug Info (remove in production) -->
    <div v-if="false" class="debug-info">
      <h4>Debug Information</h4>
      <p>Has Data: {{ hasData }}</p>
      <p>Has Valid Data: {{ hasValidData }}</p>
      <p>Raw SBOM Data Type: {{ props.sbomData ? typeof props.sbomData : 'undefined' }}</p>
      <p>Parsed SBOM Data: {{ parsedSbomData ? 'Available' : 'Not Available' }}</p>
      <p>Components: {{ parsedSbomData?.components?.length || 0 }}</p>
      <p>First Component: {{ parsedSbomData?.components?.[0] ? JSON.stringify(parsedSbomData.components[0]).substring(0, 100) + '...' : 'None' }}</p>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Loading analytics data...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="errorMessage" class="error-container">
      <div class="error-icon">⚠️</div>
      <h3>Error Loading Data</h3>
      <p>{{ errorMessage }}</p>
      <button @click="refreshData" class="retry-button">Try Again</button>
    </div>

    <!-- Empty State -->
    <div v-else-if="!hasData" class="empty-state">
      <div class="empty-icon">📊</div>
      <h3>No Analytics Data Available</h3>
      <p>Generate an SBOM first to view analytics insights</p>
      <button @click="navigateToGenerate" class="primary-button">Generate SBOM</button>
    </div>

    <!-- Analytics Dashboard -->
    <div v-else class="analytics-dashboard">
      <!-- Key Metrics Summary -->
      <div class="metrics-summary">
        <div class="metric-card total-packages">
          <div class="metric-icon">📦</div>
          <div class="metric-content">
            <div class="metric-value">{{ totalPackages }}</div>
            <div class="metric-label">Total Packages</div>
          </div>
          <div class="metric-trend" :class="packagesTrend.direction">
            {{ packagesTrend.value }}% {{ packagesTrend.direction === 'up' ? '↑' : '↓' }}
          </div>
        </div>

        <div class="metric-card vulnerabilities">
          <div class="metric-icon">🛡️</div>
          <div class="metric-content">
            <div class="metric-value">{{ totalVulnerabilities }}</div>
            <div class="metric-label">Vulnerabilities</div>
          </div>
          <div class="metric-trend" :class="vulnerabilitiesTrend.direction">
            {{ vulnerabilitiesTrend.value }}% {{ vulnerabilitiesTrend.direction === 'up' ? '↑' : '↓' }}
          </div>
        </div>

        <div class="metric-card licenses">
          <div class="metric-icon">📜</div>
          <div class="metric-content">
            <div class="metric-value">{{ uniqueLicenses }}</div>
            <div class="metric-label">Unique Licenses</div>
          </div>
          <div class="metric-trend" :class="licensesTrend.direction">
            {{ licensesTrend.value }}% {{ licensesTrend.direction === 'up' ? '↑' : '↓' }}
          </div>
        </div>

        <div class="metric-card health-score">
          <div class="metric-icon">🏆</div>
          <div class="metric-content">
            <div class="metric-value">{{ healthScore }}<span class="percentage">%</span></div>
            <div class="metric-label">Health Score</div>
          </div>
          <div class="metric-trend" :class="healthTrend.direction">
            {{ healthTrend.value }}% {{ healthTrend.direction === 'up' ? '↑' : '↓' }}
          </div>
        </div>
      </div>

      <!-- Main Analytics Content -->
      <div class="analytics-content">
        <!-- Left Column -->
        <div class="analytics-column">
          <!-- Vulnerability Breakdown -->
          <div class="analytics-card vulnerability-breakdown">
            <div class="card-header">
              <h3>Vulnerability Severity Breakdown</h3>
              <div class="card-actions">
                <button @click="exportVulnerabilityData" class="action-button">
                  Export
                </button>
              </div>
            </div>
            <div class="card-body">
              <div class="severity-chart">
                <div v-for="(severity, index) in vulnerabilitySeverity" :key="index"
                     class="severity-bar"
                     :style="{ width: severity.percentage + '%', backgroundColor: severity.color }">
                  <span class="severity-label">{{ severity.name }}</span>
                  <span class="severity-count">{{ severity.count }}</span>
                </div>
              </div>
              <div class="severity-legend">
                <div v-for="(severity, index) in vulnerabilitySeverity" :key="'legend-'+index" class="legend-item">
                  <span class="legend-color" :style="{ backgroundColor: severity.color }"></span>
                  <span class="legend-label">{{ severity.name }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Vulnerability Trend -->
          <vulnerability-trend v-if="hasValidData" :sbom-data="parsedSbomData" />
          <div v-else class="component-loading">
            <div class="loading-spinner"></div>
            <p>Loading vulnerability trend data...</p>
          </div>
        </div>

        <!-- Right Column -->
        <div class="analytics-column">
          <!-- License Compliance -->
          <license-compliance v-if="hasValidData" :sbom-data="parsedSbomData" />
          <div v-else class="component-loading">
            <div class="loading-spinner"></div>
            <p>Loading license compliance data...</p>
          </div>

          <!-- Dependency Graph -->
          <dependency-graph v-if="hasValidData" :sbom-data="parsedSbomData" />
          <div v-else class="component-loading">
            <div class="loading-spinner"></div>
            <p>Loading dependency graph data...</p>
          </div>
        </div>
      </div>

      <!-- Package Metrics Section -->
      <div class="package-metrics-section">
        <div class="section-header">
          <h3>Package Metrics</h3>
          <p class="section-description">Analyze popularity and maintenance metrics for packages in your SBOM</p>
        </div>
        <!-- Only render PackageMetrics if we have valid data -->
        <package-metrics v-if="hasValidData" :sbom-data="parsedSbomData" />
        <!-- Show loading state if we don't have data yet -->
        <div v-else class="component-loading">
          <div class="loading-spinner"></div>
          <p>Loading package metrics...</p>
        </div>
      </div>

      <!-- Recommendations Section -->
      <div class="recommendations-section">
        <div class="section-header">
          <h3>Recommendations</h3>
          <p class="section-description">Actionable insights to improve your SBOM health</p>
        </div>
        <div class="recommendations-grid">
          <div v-for="(recommendation, index) in recommendations" :key="index" class="recommendation-card">
            <div class="recommendation-header" :class="recommendation.priority">
              <div class="recommendation-icon">{{ recommendation.icon }}</div>
              <div class="recommendation-priority">{{ recommendation.priority }}</div>
            </div>
            <div class="recommendation-content">
              <h4>{{ recommendation.title }}</h4>
              <p>{{ recommendation.description }}</p>
            </div>
            <div class="recommendation-actions">
              <button class="action-button">Implement</button>
              <button class="action-button secondary">Dismiss</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue';
import VulnerabilityTrend from './VulnerabilityTrend.vue';
import PackageMetrics from './PackageMetrics.vue';
import DependencyGraph from './DependencyGraph.vue';
import LicenseCompliance from './LicenseCompliance.vue';

export default {
  name: 'AnalyticsView',
  components: {
    VulnerabilityTrend,
    PackageMetrics,
    DependencyGraph,
    LicenseCompliance
  },
  props: {
    sbomData: {
      type: Object,
      required: true
    }
  },
  setup(props, { emit }) {
    const isLoading = ref(false);
    const errorMessage = ref(null);
    const selectedTimeRange = ref('30');

    // Parse sbomData if it's a string
    const parsedSbomData = computed(() => {
      try {
        if (!props.sbomData) return null;

        // If it's a string, try to parse it
        if (typeof props.sbomData === 'string') {
          return JSON.parse(props.sbomData);
        }

        // If it's already an object, return it
        return props.sbomData;
      } catch (error) {
        console.error('Error parsing SBOM data in AnalyticsView:', error);
        return null;
      }
    });

    // Computed property to check if we have data
    const hasData = computed(() => {
      const data = parsedSbomData.value;
      return data && Array.isArray(data.components) && data.components.length > 0;
    });

    // More strict validation for components that need complete data
    const hasValidData = computed(() => {
      // Use the parsed data
      const data = parsedSbomData.value;

      // Check if data exists and has components array
      if (!data || !Array.isArray(data.components)) {
        return false;
      }

      // Check if components array has at least one item
      if (data.components.length === 0) {
        return false;
      }

      // Check if the first component has the expected properties
      const firstComponent = data.components[0];
      return firstComponent && typeof firstComponent === 'object';
    });

    // Key metrics
    const totalPackages = computed(() => {
      if (!hasData.value || !parsedSbomData.value) return 0;
      return parsedSbomData.value.components.length;
    });

    // Generate random trend data (replace with actual data in production)
    const generateTrendData = (min, max) => {
      const value = Math.floor(Math.random() * (max - min + 1)) + min;
      return {
        value,
        direction: Math.random() > 0.5 ? 'up' : 'down'
      };
    };

    // Trend data for metrics
    const packagesTrend = ref(generateTrendData(1, 15));
    const vulnerabilitiesTrend = ref(generateTrendData(5, 20));
    const licensesTrend = ref(generateTrendData(1, 10));
    const healthTrend = ref(generateTrendData(1, 8));

    // Vulnerability metrics
    const totalVulnerabilities = computed(() => {
      if (!hasData.value) return 0;

      // In a real implementation, you would calculate this from actual vulnerability data
      // For now, we'll generate a random number based on the number of components
      return Math.floor(totalPackages.value * 0.3);
    });

    // Vulnerability severity breakdown
    const vulnerabilitySeverity = computed(() => {
      if (!hasData.value) return [];

      // In a real implementation, you would calculate this from actual vulnerability data
      const total = totalVulnerabilities.value;

      return [
        {
          name: 'Critical',
          count: Math.floor(total * 0.15),
          percentage: 15,
          color: '#DC2626' // Red
        },
        {
          name: 'High',
          count: Math.floor(total * 0.25),
          percentage: 25,
          color: '#F97316' // Orange
        },
        {
          name: 'Medium',
          count: Math.floor(total * 0.35),
          percentage: 35,
          color: '#F59E0B' // Amber
        },
        {
          name: 'Low',
          count: Math.floor(total * 0.25),
          percentage: 25,
          color: '#10B981' // Green
        }
      ];
    });

    // License metrics
    const uniqueLicenses = computed(() => {
      if (!hasData.value) return 0;

      // In a real implementation, you would calculate this from actual license data
      // For now, we'll generate a random number based on the number of components
      return Math.floor(totalPackages.value * 0.2);
    });

    // Health score
    const healthScore = computed(() => {
      if (!hasData.value) return 0;

      // In a real implementation, you would calculate this based on various factors
      // For now, we'll generate a random score between 60 and 95
      return Math.floor(Math.random() * 35) + 60;
    });

    // Recommendations
    const recommendations = ref([
      {
        icon: '⚠️',
        priority: 'high',
        title: 'Update 3 Critical Packages',
        description: 'Three packages have critical vulnerabilities that should be addressed immediately.'
      },
      {
        icon: '📜',
        priority: 'medium',
        title: 'License Compliance Issues',
        description: 'Two packages have licenses that may conflict with your project\'s license.'
      },
      {
        icon: '🔄',
        priority: 'medium',
        title: 'Outdated Dependencies',
        description: '8 packages are significantly outdated and should be updated.'
      },
      {
        icon: '📦',
        priority: 'low',
        title: 'Duplicate Dependencies',
        description: 'Your project contains 3 duplicate dependencies that could be consolidated.'
      }
    ]);

    // Methods
    function refreshData() {
      isLoading.value = true;
      errorMessage.value = null;

      // Simulate API call
      setTimeout(() => {
        // Regenerate trend data
        packagesTrend.value = generateTrendData(1, 15);
        vulnerabilitiesTrend.value = generateTrendData(5, 20);
        licensesTrend.value = generateTrendData(1, 10);
        healthTrend.value = generateTrendData(1, 8);

        isLoading.value = false;
      }, 1000);
    }

    function navigateToGenerate() {
      // Emit an event that the parent component can listen for
      emit('navigate', 'generate');

      // Or use direct DOM manipulation to scroll to the generate section
      const generateSection = document.getElementById('generate');
      if (generateSection) {
        generateSection.scrollIntoView({ behavior: 'smooth' });
      }
    }

    function exportVulnerabilityData() {
      // In a real implementation, this would export the data to CSV or JSON
      alert('Exporting vulnerability data...');
    }

    // Initialize data
    onMounted(() => {
      refreshData();
    });

    return {
      isLoading,
      errorMessage,
      selectedTimeRange,
      hasData,
      hasValidData,
      parsedSbomData,
      totalPackages,
      totalVulnerabilities,
      uniqueLicenses,
      healthScore,
      packagesTrend,
      vulnerabilitiesTrend,
      licensesTrend,
      healthTrend,
      vulnerabilitySeverity,
      recommendations,
      refreshData,
      navigateToGenerate,
      exportVulnerabilityData
    };
  }
};
</script>

<style scoped>
.analytics-view {
  width: 100%;
}

.debug-info {
  margin-bottom: 1rem;
  padding: 1rem;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  font-size: 0.875rem;
}

.debug-info h4 {
  margin-top: 0;
  margin-bottom: 0.5rem;
  color: #0f766e;
}

.debug-info p {
  margin: 0.25rem 0;
  font-family: monospace;
}

.analytics-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.header-content h2 {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--dark-color);
  margin-bottom: 0.5rem;
}

.description {
  color: var(--secondary-color);
  font-size: 1rem;
}

.header-actions {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.time-range-selector {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.time-range-selector label {
  font-size: 0.875rem;
  color: var(--secondary-color);
}

.refresh-button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background-color: var(--light-color);
  border: 1px solid var(--border-color);
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--dark-color);
  cursor: pointer;
  transition: all 0.2s ease;
}

.refresh-button:hover {
  background-color: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

.refresh-icon {
  font-size: 1rem;
}

/* Loading State */
.loading-container, .component-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  background-color: var(--light-color);
  border-radius: 0.5rem;
  text-align: center;
}

.component-loading {
  padding: 2rem;
  min-height: 200px;
}

.loading-spinner {
  width: 3rem;
  height: 3rem;
  border: 0.25rem solid rgba(13, 148, 136, 0.2);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Error State */
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  background-color: rgba(220, 38, 38, 0.05);
  border-radius: 0.5rem;
  text-align: center;
}

.error-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.retry-button {
  margin-top: 1.5rem;
  padding: 0.75rem 1.5rem;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.retry-button:hover {
  background-color: var(--primary-hover);
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  background-color: var(--light-color);
  border-radius: 0.5rem;
  text-align: center;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.7;
}

.primary-button {
  margin-top: 1.5rem;
  padding: 0.75rem 1.5rem;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.primary-button:hover {
  background-color: var(--primary-hover);
}

/* Analytics Dashboard */
.analytics-dashboard {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

/* Metrics Summary */
.metrics-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.5rem;
  margin-bottom: 1rem;
}

.metric-card {
  display: flex;
  align-items: center;
  padding: 1.5rem;
  background-color: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.metric-icon {
  font-size: 2rem;
  margin-right: 1rem;
}

.metric-content {
  flex: 1;
}

.metric-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--dark-color);
  line-height: 1.2;
}

.percentage {
  font-size: 1rem;
  font-weight: 500;
  color: var(--secondary-color);
}

.metric-label {
  font-size: 0.875rem;
  color: var(--secondary-color);
}

.metric-trend {
  font-size: 0.875rem;
  font-weight: 500;
  padding: 0.25rem 0.5rem;
  border-radius: 9999px;
}

.metric-trend.up {
  background-color: rgba(16, 185, 129, 0.1);
  color: rgb(16, 185, 129);
}

.metric-trend.down {
  background-color: rgba(220, 38, 38, 0.1);
  color: rgb(220, 38, 38);
}

/* Analytics Content */
.analytics-content {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
  gap: 1.5rem;
}

.analytics-column {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.analytics-card {
  background-color: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.card-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--dark-color);
  margin: 0;
}

.card-actions {
  display: flex;
  gap: 0.5rem;
}

.action-button {
  padding: 0.375rem 0.75rem;
  background-color: var(--light-color);
  border: 1px solid var(--border-color);
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--dark-color);
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-button:hover {
  background-color: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

.action-button.secondary:hover {
  background-color: var(--secondary-color);
  border-color: var(--secondary-color);
}

.card-body {
  padding: 1.5rem;
}

/* Vulnerability Breakdown */
.severity-chart {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.severity-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  color: white;
  font-size: 0.875rem;
  font-weight: 500;
  min-width: 30px;
  transition: width 0.5s ease;
}

.severity-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-top: 1rem;
}

.legend-item {
  display: flex;
  align-items: center;
  font-size: 0.75rem;
  color: var(--secondary-color);
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  margin-right: 0.5rem;
}

/* Package Metrics Section */
.package-metrics-section {
  margin-top: 1rem;
}

.section-header {
  margin-bottom: 1rem;
}

.section-header h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--dark-color);
  margin-bottom: 0.5rem;
}

.section-description {
  font-size: 0.875rem;
  color: var(--secondary-color);
}

/* Recommendations Section */
.recommendations-section {
  margin-top: 2rem;
}

.recommendations-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
}

.recommendation-card {
  background-color: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.recommendation-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.recommendation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.recommendation-header.high {
  background-color: rgba(220, 38, 38, 0.1);
}

.recommendation-header.medium {
  background-color: rgba(245, 158, 11, 0.1);
}

.recommendation-header.low {
  background-color: rgba(16, 185, 129, 0.1);
}

.recommendation-icon {
  font-size: 1.25rem;
}

.recommendation-priority {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  padding: 0.25rem 0.5rem;
  border-radius: 9999px;
  background-color: white;
}

.recommendation-header.high .recommendation-priority {
  color: rgb(220, 38, 38);
}

.recommendation-header.medium .recommendation-priority {
  color: rgb(245, 158, 11);
}

.recommendation-header.low .recommendation-priority {
  color: rgb(16, 185, 129);
}

.recommendation-content {
  padding: 1.5rem;
}

.recommendation-content h4 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--dark-color);
  margin-top: 0;
  margin-bottom: 0.5rem;
}

.recommendation-content p {
  font-size: 0.875rem;
  color: var(--secondary-color);
  margin: 0;
}

.recommendation-actions {
  display: flex;
  gap: 0.5rem;
  padding: 0 1.5rem 1.5rem;
}

/* Responsive Design */
@media (max-width: 768px) {
  .analytics-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    flex-direction: column;
    align-items: flex-start;
  }

  .time-range-selector {
    width: 100%;
  }

  .time-range-selector select {
    width: 100%;
  }

  .refresh-button {
    width: 100%;
    justify-content: center;
  }

  .analytics-content {
    grid-template-columns: 1fr;
  }

  .metrics-summary {
    grid-template-columns: 1fr;
  }

  .recommendations-grid {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 769px) and (max-width: 1200px) {
  .analytics-content {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 1201px) {
  .analytics-view {
    max-width: 1400px;
    margin: 0 auto;
  }
}
</style>
