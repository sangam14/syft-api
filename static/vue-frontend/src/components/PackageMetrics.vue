<template>
  <div class="package-metrics">
    <div class="metrics-header">
      <div class="metrics-title">
        <h3>Package Metrics</h3>
        <span class="package-count" v-if="!isLoading && displayedPackages && Array.isArray(displayedPackages)">{{ displayedPackages.length || 0 }} / {{ totalPackages || 0 }} packages</span>
      </div>
      <div class="metrics-filters">
        <div class="filter-group">
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
        <div class="range-selector">
          <label for="packageRange">Show:</label>
          <div class="range-inputs">
            <input
              type="number"
              id="packageRange"
              v-model.number="displayLimit"
              min="5"
              :max="filteredPackages.length"
              @change="validateDisplayLimit"
              class="modern-input range-input"
            />
            <div class="range-controls">
              <button
                @click="increaseRange"
                class="range-button"
                :disabled="displayLimit >= filteredPackages.length"
                title="Show more packages"
              >+</button>
              <button
                @click="decreaseRange"
                class="range-button"
                :disabled="displayLimit <= 5"
                title="Show fewer packages"
              >-</button>
            </div>
          </div>
          <span class="range-info" v-if="filteredPackages.length > 0">
            of {{ filteredPackages.length }}
          </span>
        </div>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>Loading package metrics...</p>
    </div>

    <!-- Empty state when no packages match the filter -->
    <div v-else-if="filteredPackages.length === 0" class="empty-state">
      <div class="empty-icon">📦</div>
      <p v-if="searchQuery">No packages match your search criteria "{{ searchQuery }}".</p>
      <p v-else>No package data available.</p>
      <button
        v-if="searchQuery && packages.length > 0"
        @click="searchQuery = ''"
        class="reset-button"
      >
        Clear Search
      </button>
    </div>

    <!-- Package summary section -->
    <div v-else-if="filteredPackages.length > 0" class="metrics-content">
      <div class="package-summary" v-if="filteredPackages.length > 1">
        <div class="summary-item">
          <div class="summary-label">Total Packages</div>
          <div class="summary-value">{{ filteredPackages.length }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">Avg. Downloads</div>
          <div class="summary-value">{{ formatNumber(averageDownloads) }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">Avg. Stars</div>
          <div class="summary-value">{{ formatNumber(averageStars) }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">Avg. Maintenance</div>
          <div class="summary-value" :class="getMaintenanceClass(averageMaintenance)">{{ averageMaintenance }}%</div>
        </div>
      </div>

      <!-- Package metrics grid -->
      <div class="metrics-grid">
        <div v-if="!displayedPackages || !Array.isArray(displayedPackages) || displayedPackages.length === 0" class="empty-metrics">
          <p>No package metrics available to display.</p>
        </div>
        <div v-else v-for="pkg in displayedPackages" :key="pkg?.name || 'unknown'" class="metric-card">
          <div class="metric-header">
            <h4>{{ pkg?.name || 'Unknown Package' }}</h4>
            <span class="metric-badge" :class="getMaintenanceClass(pkg?.maintenanceScore || 0)">
              {{ pkg?.maintenanceScore || 0 }}%
            </span>
          </div>

          <div class="metric-details">
            <div class="metric-row">
              <span class="metric-label">Downloads</span>
              <span class="metric-value">{{ formatNumber(pkg?.downloads || 0) }}</span>
            </div>
            <div class="metric-row">
              <span class="metric-label">Stars</span>
              <span class="metric-value">{{ formatNumber(pkg?.stars || 0) }}</span>
            </div>
            <div class="metric-row">
              <span class="metric-label">Open Issues</span>
              <span class="metric-value">{{ pkg?.issues || 0 }}</span>
            </div>
            <div class="metric-row">
              <span class="metric-label">Last Updated</span>
              <span class="metric-value">{{ formatDate(pkg?.lastUpdated || new Date().toISOString()) }}</span>
            </div>
          </div>

          <div class="metric-progress">
            <div class="progress-bar">
              <div
                class="progress-fill"
                :style="{ width: `${pkg?.maintenanceScore || 0}%` }"
                :class="getMaintenanceClass(pkg?.maintenanceScore || 0)"
              ></div>
            </div>
            <span class="progress-label">Maintenance Score</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue';

export default {
  name: 'PackageMetrics',
  props: {
    sbomData: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    // Initialize with defensive defaults
    const searchQuery = ref('');
    const selectedMetric = ref('downloads');
    const isLoading = ref(true); // Start with loading state
    const packageCache = ref(new Map());
    const debouncedSearchQuery = ref('');
    const displayLimit = ref(10); // Default number of packages to display

    // Initialize a safe empty array for packages
    const safePackages = ref([]);

    // Debounce search input to prevent excessive filtering
    let searchTimeout = null;

    // Watch packages and update safePackages when packages changes
    watch(packages, (newPackages) => {
      try {
        if (newPackages && Array.isArray(newPackages) && newPackages.length > 0) {
          // Only update if we have actual data (not just the loading placeholder)
          if (!(newPackages.length === 1 && newPackages[0]?.name === 'Loading...')) {
            safePackages.value = [...newPackages];
            console.log('Updated safePackages with new package data');
          }
        }
      } catch (error) {
        console.error('Error updating safePackages:', error);
      }
    }, { immediate: false });

    watch(searchQuery, (newValue) => {
      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }

      searchTimeout = setTimeout(() => {
        debouncedSearchQuery.value = newValue;
      }, 300);
    });

    // Watch for changes in filtered packages to adjust display limit
    // Use { immediate: false } to prevent the watcher from running immediately
    watch(filteredPackages, (newPackages) => {
      try {
        // Ensure newPackages is an array
        if (!newPackages) {
          console.log('newPackages is undefined in filteredPackages watcher');
          displayLimit.value = 10; // Reset to default
          return;
        }

        if (!Array.isArray(newPackages)) {
          console.log('newPackages is not an array in filteredPackages watcher');
          displayLimit.value = 10; // Reset to default
          return;
        }

        // If current display limit is greater than available packages, adjust it
        if (displayLimit.value > newPackages.length) {
          displayLimit.value = Math.max(5, newPackages.length);
        }

        // If we have no packages, reset to default
        if (newPackages.length === 0) {
          displayLimit.value = 10; // Reset to default for next time
        }
      } catch (error) {
        console.error('Error in filteredPackages watcher:', error);
        // Set a safe default
        displayLimit.value = 10;
      }
    }, { immediate: false });

    // Generate deterministic data based on component properties
    const generatePackageData = (component) => {
      // Create a simple hash from the component name for deterministic values
      const nameHash = component.name.split('').reduce((acc, char) => {
        return acc + char.charCodeAt(0);
      }, 0);

      // Use the hash to generate deterministic values
      const downloads = (nameHash * 1000) % 10000000;
      const stars = (nameHash * 100) % 50000;
      const issues = (nameHash * 10) % 50;
      const maintenanceScore = (nameHash * 5) % 100;

      // Generate a deterministic date within the last year
      const today = new Date();
      const daysAgo = (nameHash % 365) + 1;
      const lastUpdated = new Date(today);
      lastUpdated.setDate(lastUpdated.getDate() - daysAgo);

      return {
        name: component.name,
        version: component.version,
        downloads: Math.floor(downloads),
        stars: Math.floor(stars),
        issues: Math.floor(issues),
        lastUpdated: lastUpdated.toISOString().split('T')[0],
        maintenanceScore: Math.floor(maintenanceScore)
      };
    };

    // Use memoization to avoid recalculating package data
    const getPackageData = (component) => {
      const cacheKey = `${component.name}-${component.version}`;

      if (!packageCache.value.has(cacheKey)) {
        packageCache.value.set(cacheKey, generatePackageData(component));
      }

      return packageCache.value.get(cacheKey);
    };

    // Optimized computed property with memoization
    const packages = computed(() => {
      try {
        // Check if sbomData exists and has components array
        if (!props.sbomData) {
          console.log('No sbomData provided to PackageMetrics');
          return safePackages.value;
        }

        // Handle case where sbomData might be a string
        let componentsArray = [];

        if (typeof props.sbomData === 'string') {
          try {
            const parsed = JSON.parse(props.sbomData);
            componentsArray = parsed.components || [];
          } catch (parseError) {
            console.error('Error parsing sbomData string:', parseError);
            return safePackages.value;
          }
        } else if (props.sbomData.components && Array.isArray(props.sbomData.components)) {
          componentsArray = props.sbomData.components;
        } else {
          console.log('sbomData does not contain a valid components array');
          return safePackages.value;
        }

        // Map components to package data
        const result = componentsArray.map(component => {
          // Ensure component has required properties
          if (!component || !component.name) {
            return {
              name: 'Unknown',
              version: 'Unknown',
              downloads: 0,
              stars: 0,
              issues: 0,
              lastUpdated: new Date().toISOString().split('T')[0],
              maintenanceScore: 0
            };
          }
          return getPackageData(component);
        });

        // We'll update safePackages in a watcher instead of here to avoid side effects
        return result;
      } catch (error) {
        console.error('Error in packages computed property:', error);
        return safePackages.value;
      }
    });

    // Total number of packages
    const totalPackages = computed(() => packages.value ? packages.value.length : 0);

    // Optimized filtering and sorting with memoization
    const filteredPackages = computed(() => {
      try {
        // Ensure packages exists and is an array
        if (!packages.value || !Array.isArray(packages.value)) {
          console.log('packages is not an array in filteredPackages');
          return safePackages.value;
        }

        // Get query and metric with fallbacks
        const query = (debouncedSearchQuery.value || '').toLowerCase();
        const metric = selectedMetric.value || 'downloads';

        // Create a safe copy of the packages array
        const safePackages = [...packages.value];

        // Filter packages based on search query
        const filtered = query
          ? safePackages.filter(pkg => {
              // Ensure pkg and pkg.name exist
              if (!pkg || typeof pkg.name !== 'string') return false;
              return pkg.name.toLowerCase().includes(query);
            })
          : safePackages;

        // Sort packages based on selected metric
        return filtered.sort((a, b) => {
          // Handle null or undefined values
          if (!a && !b) return 0;
          if (!a) return 1; // b comes first
          if (!b) return -1; // a comes first

          switch (metric) {
            case 'downloads':
              return (b.downloads || 0) - (a.downloads || 0);
            case 'stars':
              return (b.stars || 0) - (a.stars || 0);
            case 'issues':
              return (a.issues || 0) - (b.issues || 0); // Lower is better for issues
            case 'maintenance':
              return (b.maintenanceScore || 0) - (a.maintenanceScore || 0);
            default:
              return 0;
          }
        });
      } catch (error) {
        console.error('Error in filteredPackages computed property:', error);
        return safePackages.value;
      }
    });


    // Calculate average metrics for summary
    const averageDownloads = computed(() => {
      try {
        // Use filteredPackages if available, otherwise fall back to safePackages
        const packagesToUse = (filteredPackages.value && Array.isArray(filteredPackages.value) && filteredPackages.value.length > 0)
          ? filteredPackages.value
          : safePackages.value;

        if (packagesToUse.length === 0) return 0;

        const sum = packagesToUse.reduce((acc, pkg) => acc + (pkg?.downloads || 0), 0);
        return Math.round(sum / packagesToUse.length);
      } catch (error) {
        console.error('Error in averageDownloads computed property:', error);
        return 0;
      }
    });

    const averageStars = computed(() => {
      try {
        // Use filteredPackages if available, otherwise fall back to safePackages
        const packagesToUse = (filteredPackages.value && Array.isArray(filteredPackages.value) && filteredPackages.value.length > 0)
          ? filteredPackages.value
          : safePackages.value;

        if (packagesToUse.length === 0) return 0;

        const sum = packagesToUse.reduce((acc, pkg) => acc + (pkg?.stars || 0), 0);
        return Math.round(sum / packagesToUse.length);
      } catch (error) {
        console.error('Error in averageStars computed property:', error);
        return 0;
      }
    });

    const averageMaintenance = computed(() => {
      try {
        // Use filteredPackages if available, otherwise fall back to safePackages
        const packagesToUse = (filteredPackages.value && Array.isArray(filteredPackages.value) && filteredPackages.value.length > 0)
          ? filteredPackages.value
          : safePackages.value;

        if (packagesToUse.length === 0) return 0;

        const sum = packagesToUse.reduce((acc, pkg) => acc + (pkg?.maintenanceScore || 0), 0);
        return Math.round(sum / packagesToUse.length);
      } catch (error) {
        console.error('Error in averageMaintenance computed property:', error);
        return 0;
      }
    });

    // Watch for changes in display limit to validate it
    watch(displayLimit, () => {
      try {
        validateDisplayLimit();
      } catch (error) {
        console.error('Error in displayLimit watcher:', error);
        // Set a safe default
        displayLimit.value = 10;
      }
    }, { immediate: false });

    // Displayed packages with limit applied
    const displayedPackages = computed(() => {
      try {
        // Ensure filteredPackages exists and has a length property
        if (!filteredPackages.value) {
          console.log('filteredPackages is undefined in displayedPackages');
          return safePackages.value.slice(0, 5); // Return at most 5 items from safe packages
        }

        if (!Array.isArray(filteredPackages.value)) {
          console.log('filteredPackages is not an array in displayedPackages');
          return safePackages.value.slice(0, 5); // Return at most 5 items from safe packages
        }

        if (filteredPackages.value.length === 0) {
          return safePackages.value.length > 0 ? safePackages.value.slice(0, 5) : [];
        }

        // Ensure displayLimit is a valid number
        const numLimit = Number(displayLimit.value);
        const validLimit = isNaN(numLimit) ? 5 : numLimit;
        const limit = Math.max(5, Math.min(filteredPackages.value.length, validLimit));

        // If the limit needs correction, we'll handle it in the next tick via the watcher
        // This avoids modifying state in computed properties

        return filteredPackages.value.slice(0, limit);
      } catch (error) {
        console.error('Error in displayedPackages computed property:', error);
        return safePackages.value.slice(0, 5); // Return at most 5 items from safe packages
      }
    });

    // Functions to increase/decrease the display limit
    function increaseRange() {
      try {
        // Safely get the maximum available packages
        let maxPackages = 0;

        if (filteredPackages.value && Array.isArray(filteredPackages.value)) {
          maxPackages = filteredPackages.value.length;
        } else {
          console.log('filteredPackages is not valid in increaseRange');
          // Default to a reasonable value if we can't determine the actual max
          maxPackages = 20;
        }

        // Get current display limit as a number with fallback
        const currentLimit = Number(displayLimit.value) || 5;

        // Increase by 5, but don't exceed maxPackages
        displayLimit.value = Math.min(maxPackages, currentLimit + 5);
      } catch (error) {
        console.error('Error in increaseRange function:', error);
      }
    }

    function decreaseRange() {
      try {
        // Get current display limit as a number with fallback
        const currentLimit = Number(displayLimit.value) || 10;

        // Decrease by 5, but don't go below 5
        displayLimit.value = Math.max(5, currentLimit - 5);
      } catch (error) {
        console.error('Error in decreaseRange function:', error);
        // Set a safe default
        displayLimit.value = 5;
      }
    }

    // Validate and correct the display limit if needed
    function validateDisplayLimit() {
      try {
        // Convert to number in case it's a string from the input
        const numValue = Number(displayLimit.value);

        // Handle NaN or invalid values
        if (isNaN(numValue) || numValue < 5) {
          displayLimit.value = 5;
          return;
        }

        // Safely get the maximum available packages
        let maxPackages = 0;

        if (filteredPackages.value && Array.isArray(filteredPackages.value)) {
          maxPackages = filteredPackages.value.length;
        } else {
          console.log('filteredPackages is not valid in validateDisplayLimit');
          // Default to 10 if we can't determine the actual max
          maxPackages = 10;
        }

        if (numValue > maxPackages && maxPackages > 0) {
          displayLimit.value = maxPackages;
          return;
        }

        // Ensure it's an integer
        displayLimit.value = Math.floor(numValue);
      } catch (error) {
        console.error('Error in validateDisplayLimit function:', error);
        // Set a safe default
        displayLimit.value = 5;
      }
    }

    // Optimized number formatter with memoization
    const numberFormatCache = new Map();

    const formatNumber = (num) => {
      try {
        // Handle undefined, null, or NaN
        if (num === undefined || num === null || isNaN(num)) {
          return '0';
        }

        // Ensure num is a number
        const numValue = Number(num);

        // Check cache first
        if (numberFormatCache.has(numValue)) {
          return numberFormatCache.get(numValue);
        }

        let result;
        if (numValue >= 1000000) {
          result = (numValue / 1000000).toFixed(1) + 'M';
        } else if (numValue >= 1000) {
          result = (numValue / 1000).toFixed(1) + 'K';
        } else {
          result = numValue.toString();
        }

        // Cache the result
        numberFormatCache.set(numValue, result);
        return result;
      } catch (error) {
        console.error('Error in formatNumber function:', error);
        return '0';
      }
    };

    // Optimized date formatter with memoization
    const dateFormatCache = new Map();

    const formatDate = (dateString) => {
      try {
        // Handle undefined or null
        if (!dateString) {
          return 'Unknown date';
        }

        // Check cache first
        if (dateFormatCache.has(dateString)) {
          return dateFormatCache.get(dateString);
        }

        // Try to create a valid date
        const date = new Date(dateString);

        // Check if date is valid
        if (isNaN(date.getTime())) {
          return 'Invalid date';
        }

        const result = date.toLocaleDateString('en-US', {
          year: 'numeric',
          month: 'short',
          day: 'numeric'
        });

        // Cache the result
        dateFormatCache.set(dateString, result);
        return result;
      } catch (error) {
        console.error('Error in formatDate function:', error);
        return 'Unknown date';
      }
    };

    // Optimized class getter with memoization
    const maintenanceClassCache = new Map();

    const getMaintenanceClass = (score) => {
      try {
        // Handle undefined, null, or NaN
        if (score === undefined || score === null || isNaN(score)) {
          return 'poor';
        }

        // Ensure score is a number
        const numScore = Number(score);

        // Check cache first
        if (maintenanceClassCache.has(numScore)) {
          return maintenanceClassCache.get(numScore);
        }

        let result;
        if (numScore >= 90) result = 'excellent';
        else if (numScore >= 75) result = 'good';
        else if (numScore >= 60) result = 'fair';
        else result = 'poor';

        // Cache the result
        maintenanceClassCache.set(numScore, result);
        return result;
      } catch (error) {
        console.error('Error in getMaintenanceClass function:', error);
        return 'poor';
      }
    };

    // Initialize component and simulate loading state for better UX
    onMounted(() => {
      // Initialize with some safe default data
      safePackages.value = [
        {
          name: 'Loading...',
          version: '1.0.0',
          downloads: 0,
          stars: 0,
          issues: 0,
          lastUpdated: new Date().toISOString().split('T')[0],
          maintenanceScore: 50
        }
      ];

      // Simulate loading state
      isLoading.value = true;
      setTimeout(() => {
        // If we still don't have real data, keep the loading state
        if (packages.value.length <= 1 && packages.value[0]?.name === 'Loading...') {
          console.log('No real package data available after timeout');
        } else {
          isLoading.value = false;
        }
      }, 1000);
    });

    return {
      searchQuery,
      selectedMetric,
      filteredPackages,
      displayedPackages,
      totalPackages,
      displayLimit,
      increaseRange,
      decreaseRange,
      validateDisplayLimit,
      formatNumber,
      formatDate,
      getMaintenanceClass,
      isLoading,
      averageDownloads,
      averageStars,
      averageMaintenance
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
  flex-direction: column;
  margin-bottom: 2rem;
}

.metrics-title {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}

.metrics-header h3 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--dark-color);
  margin-right: 1rem;
  margin-bottom: 0;
}

.package-count {
  font-size: var(--font-size-sm);
  color: var(--secondary-color);
  background-color: var(--light-color);
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
}

.metrics-filters {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.filter-group {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.metrics-filters input {
  width: 200px;
}

.range-selector {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  position: relative;
}

.range-selector label {
  font-size: var(--font-size-sm);
  color: var(--secondary-color);
  white-space: nowrap;
}

.range-inputs {
  display: flex;
  align-items: center;
}

.range-input {
  width: 70px;
  text-align: center;
  padding-right: 0.5rem;
}

.range-info {
  font-size: var(--font-size-sm);
  color: var(--secondary-color);
  white-space: nowrap;
  margin-left: 0.25rem;
}

.range-controls {
  display: flex;
  flex-direction: column;
  margin-left: 0.5rem;
}

.range-button {
  background-color: var(--light-color);
  border: 1px solid var(--border-color);
  color: var(--dark-color);
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 0;
  font-size: var(--font-size-sm);
  transition: all 0.2s ease;
}

.range-button:first-child {
  border-top-left-radius: 4px;
  border-top-right-radius: 4px;
  border-bottom: none;
}

.range-button:last-child {
  border-bottom-left-radius: 4px;
  border-bottom-right-radius: 4px;
}

.range-button:hover:not(:disabled) {
  background-color: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

.range-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .metrics-filters {
    flex-direction: column;
    align-items: flex-start;
  }

  .filter-group {
    width: 100%;
    margin-bottom: 1rem;
  }

  .filter-group input,
  .filter-group select {
    width: 100%;
    margin-bottom: 0.5rem;
  }

  .range-selector {
    width: 100%;
    justify-content: flex-start;
  }

  .metrics-title {
    flex-direction: column;
    align-items: flex-start;
  }

  .metrics-header h3 {
    margin-bottom: 0.5rem;
  }

  .package-count {
    margin-left: 0;
  }

  .metrics-grid {
    grid-template-columns: 1fr;
  }
}

.metrics-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.package-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  background-color: var(--light-color);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 0.5rem;
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.summary-label {
  font-size: var(--font-size-sm);
  color: var(--secondary-color);
  margin-bottom: 0.5rem;
}

.summary-value {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  color: var(--dark-color);
}

.summary-value.excellent {
  color: var(--success-color);
}

.summary-value.good {
  color: var(--primary-color);
}

.summary-value.fair {
  color: var(--warning-color);
}

.summary-value.poor {
  color: var(--danger-color);
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

.empty-metrics {
  padding: 2rem;
  text-align: center;
  background-color: var(--light-color);
  border-radius: 8px;
  color: var(--secondary-color);
  font-style: italic;
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

/* Loading state styles */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  background-color: var(--light-color);
  border-radius: 8px;
  text-align: center;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(13, 148, 136, 0.2);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Empty state styles */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  background-color: var(--light-color);
  border-radius: 8px;
  text-align: center;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.7;
}

.reset-button {
  margin-top: 1rem;
  padding: 0.5rem 1.5rem;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: var(--font-weight-medium);
  transition: background-color 0.2s ease;
}

.reset-button:hover {
  background-color: var(--primary-hover);
}
</style>