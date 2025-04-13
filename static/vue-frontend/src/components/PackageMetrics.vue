<template>
  <div class="package-metrics">
    <div class="metrics-header">
      <h3>Package Metrics</h3>
      <div class="metrics-filters">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="Search packages..."
          class="modern-input"
        />
        <select v-model="selectedMetric" class="modern-input">
          <option value="downloads">Downloads</option>
          <option value="stars">GitHub Stars</option>
          <option value="issues">Open Issues</option>
          <option value="maintenance">Maintenance Score</option>
        </select>
      </div>
    </div>

    <div class="metrics-grid">
      <div v-for="pkg in filteredPackages" :key="pkg.name" class="metric-card">
        <div class="metric-header">
          <h4>{{ pkg.name }}</h4>
          <span class="metric-badge" :class="getMaintenanceClass(pkg.maintenanceScore)">
            {{ pkg.maintenanceScore }}%
          </span>
        </div>
        
        <div class="metric-details">
          <div class="metric-row">
            <span class="metric-label">Downloads</span>
            <span class="metric-value">{{ formatNumber(pkg.downloads) }}</span>
          </div>
          <div class="metric-row">
            <span class="metric-label">Stars</span>
            <span class="metric-value">{{ formatNumber(pkg.stars) }}</span>
          </div>
          <div class="metric-row">
            <span class="metric-label">Open Issues</span>
            <span class="metric-value">{{ pkg.issues }}</span>
          </div>
          <div class="metric-row">
            <span class="metric-label">Last Updated</span>
            <span class="metric-value">{{ formatDate(pkg.lastUpdated) }}</span>
          </div>
        </div>

        <div class="metric-progress">
          <div class="progress-bar">
            <div 
              class="progress-fill" 
              :style="{ width: `${pkg.maintenanceScore}%` }"
              :class="getMaintenanceClass(pkg.maintenanceScore)"
            ></div>
          </div>
          <span class="progress-label">Maintenance Score</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed } from 'vue';

export default {
  name: 'PackageMetrics',
  props: {
    sbomData: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const searchQuery = ref('');
    const selectedMetric = ref('downloads');

    // Use the props.sbomData to generate package metrics
    const packages = computed(() => {
      if (!props.sbomData || !props.sbomData.components) {
        return [];
      }

      return props.sbomData.components.map(component => ({
        name: component.name,
        version: component.version,
        downloads: Math.floor(Math.random() * 10000000), // Placeholder - replace with actual data
        stars: Math.floor(Math.random() * 50000), // Placeholder - replace with actual data
        issues: Math.floor(Math.random() * 50), // Placeholder - replace with actual data
        lastUpdated: new Date().toISOString().split('T')[0], // Placeholder - replace with actual data
        maintenanceScore: Math.floor(Math.random() * 100) // Placeholder - replace with actual data
      }));
    });

    const filteredPackages = computed(() => {
      return packages.value
        .filter(pkg => 
          pkg.name.toLowerCase().includes(searchQuery.value.toLowerCase())
        )
        .sort((a, b) => {
          switch (selectedMetric.value) {
            case 'downloads':
              return b.downloads - a.downloads;
            case 'stars':
              return b.stars - a.stars;
            case 'issues':
              return b.issues - a.issues;
            case 'maintenance':
              return b.maintenanceScore - a.maintenanceScore;
            default:
              return 0;
          }
        });
    });

    const formatNumber = (num) => {
      if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
      }
      if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
      }
      return num.toString();
    };

    const formatDate = (dateString) => {
      const date = new Date(dateString);
      return date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'short', 
        day: 'numeric' 
      });
    };

    const getMaintenanceClass = (score) => {
      if (score >= 90) return 'excellent';
      if (score >= 75) return 'good';
      if (score >= 60) return 'fair';
      return 'poor';
    };

    return {
      searchQuery,
      selectedMetric,
      filteredPackages,
      formatNumber,
      formatDate,
      getMaintenanceClass
    };
  }
};
</script>

<style scoped>
.package-metrics {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: var(--card-shadow);
}

.metrics-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.metrics-header h3 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--dark-color);
}

.metrics-filters {
  display: flex;
  gap: 1rem;
}

.metrics-filters input {
  width: 200px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.metric-card {
  background: var(--light-color);
  border-radius: 8px;
  padding: 1.5rem;
  transition: transform 0.2s ease;
}

.metric-card:hover {
  transform: translateY(-2px);
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.metric-header h4 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--dark-color);
}

.metric-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.metric-badge.excellent {
  background: rgba(16, 185, 129, 0.1);
  color: rgb(16, 185, 129);
}

.metric-badge.good {
  background: rgba(37, 99, 235, 0.1);
  color: rgb(37, 99, 235);
}

.metric-badge.fair {
  background: rgba(245, 158, 11, 0.1);
  color: rgb(245, 158, 11);
}

.metric-badge.poor {
  background: rgba(220, 38, 38, 0.1);
  color: rgb(220, 38, 38);
}

.metric-details {
  margin-bottom: 1.5rem;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.metric-label {
  color: var(--secondary-color);
  font-size: var(--font-size-sm);
}

.metric-value {
  font-weight: var(--font-weight-medium);
  color: var(--dark-color);
}

.metric-progress {
  margin-top: 1rem;
}

.progress-bar {
  height: 6px;
  background: var(--light-color);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 0.5rem;
}

.progress-fill {
  height: 100%;
  transition: width 0.3s ease;
}

.progress-fill.excellent {
  background: rgb(16, 185, 129);
}

.progress-fill.good {
  background: rgb(37, 99, 235);
}

.progress-fill.fair {
  background: rgb(245, 158, 11);
}

.progress-fill.poor {
  background: rgb(220, 38, 38);
}

.progress-label {
  font-size: var(--font-size-xs);
  color: var(--secondary-color);
}
</style> 