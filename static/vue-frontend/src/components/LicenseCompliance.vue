<template>
  <div class="license-compliance">
    <div class="compliance-header">
      <h3>License Compliance</h3>
    </div>
    <div class="compliance-content">
      <div class="license-summary">
        <div class="license-chart">
          <div class="chart-placeholder">
            <div class="chart-circle" :style="{ background: `conic-gradient(var(--success-color) 0% ${compliantPercentage}%, var(--warning-color) ${compliantPercentage}% ${compliantPercentage + warningPercentage}%, var(--danger-color) ${compliantPercentage + warningPercentage}% 100%)` }">
              <div class="chart-inner">
                <span class="chart-percentage">{{ compliantPercentage }}%</span>
                <span class="chart-label">Compliant</span>
              </div>
            </div>
          </div>
          <div class="chart-legend">
            <div class="legend-item">
              <div class="legend-color compliant"></div>
              <span>Compliant ({{ compliantPercentage }}%)</span>
            </div>
            <div class="legend-item">
              <div class="legend-color warning"></div>
              <span>Review Needed ({{ warningPercentage }}%)</span>
            </div>
            <div class="legend-item">
              <div class="legend-color violation"></div>
              <span>Incompatible ({{ violationPercentage }}%)</span>
            </div>
          </div>
        </div>
      </div>
      
      <div class="license-list">
        <h4>Top Licenses</h4>
        <ul>
          <li v-for="(license, index) in topLicenses" :key="index" class="license-item">
            <div class="license-name">{{ license.name }}</div>
            <div class="license-count">{{ license.count }} packages</div>
            <div class="license-status" :class="license.status">{{ license.status }}</div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed } from 'vue';

export default {
  name: 'LicenseCompliance',
  props: {
    sbomData: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    // Compliance percentages
    const compliantPercentage = ref(75);
    const warningPercentage = ref(15);
    const violationPercentage = ref(10);

    // Calculate license data from SBOM
    const topLicenses = computed(() => {
      if (!props.sbomData || !props.sbomData.components) {
        return [
          { name: 'MIT', count: 42, status: 'compliant' },
          { name: 'Apache-2.0', count: 28, status: 'compliant' },
          { name: 'BSD-3-Clause', count: 14, status: 'compliant' },
          { name: 'GPL-3.0', count: 7, status: 'warning' },
          { name: 'LGPL-2.1', count: 5, status: 'warning' },
          { name: 'Unknown', count: 3, status: 'violation' }
        ];
      }

      // In a real implementation, process the actual licenses from props.sbomData
      // This is a placeholder
      return [
        { name: 'MIT', count: 42, status: 'compliant' },
        { name: 'Apache-2.0', count: 28, status: 'compliant' },
        { name: 'BSD-3-Clause', count: 14, status: 'compliant' },
        { name: 'GPL-3.0', count: 7, status: 'warning' },
        { name: 'LGPL-2.1', count: 5, status: 'warning' },
        { name: 'Unknown', count: 3, status: 'violation' }
      ];
    });

    return {
      compliantPercentage,
      warningPercentage,
      violationPercentage,
      topLicenses
    };
  }
};
</script>

<style scoped>
.license-compliance {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: var(--card-shadow);
}

.compliance-header {
  margin-bottom: 1.5rem;
}

.compliance-header h3 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--dark-color);
}

.compliance-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
}

.license-chart {
  text-align: center;
}

.chart-placeholder {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 1.5rem;
}

.chart-circle {
  width: 200px;
  height: 200px;
  border-radius: 50%;
  position: relative;
}

.chart-inner {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 160px;
  height: 160px;
  background: white;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.chart-percentage {
  font-size: 2.5rem;
  font-weight: var(--font-weight-bold);
  color: var(--dark-color);
}

.chart-label {
  font-size: var(--font-size-sm);
  color: var(--secondary-color);
}

.chart-legend {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  text-align: left;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.legend-color {
  width: 16px;
  height: 16px;
  border-radius: 50%;
}

.legend-color.compliant {
  background-color: var(--success-color);
}

.legend-color.warning {
  background-color: var(--warning-color);
}

.legend-color.violation {
  background-color: var(--danger-color);
}

.license-list {
  padding: 1rem;
  background: var(--light-color);
  border-radius: 8px;
}

.license-list h4 {
  margin-bottom: 1rem;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.license-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border-color);
}

.license-item:last-child {
  border-bottom: none;
}

.license-name {
  font-weight: var(--font-weight-medium);
}

.license-count {
  color: var(--secondary-color);
  font-size: var(--font-size-sm);
}

.license-status {
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  text-transform: uppercase;
}

.license-status.compliant {
  background: rgba(16, 185, 129, 0.1);
  color: rgb(16, 185, 129);
}

.license-status.warning {
  background: rgba(245, 158, 11, 0.1);
  color: rgb(245, 158, 11);
}

.license-status.violation {
  background: rgba(220, 38, 38, 0.1);
  color: rgb(220, 38, 38);
}

@media (max-width: 768px) {
  .compliance-content {
    grid-template-columns: 1fr;
  }
}
</style> 