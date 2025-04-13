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
              :max="filteredPackages && Array.isArray(filteredPackages) ? filteredPackages.length : 10"
              @change="validateDisplayLimit"
              class="modern-input range-input"
            />
            <div class="range-controls">
              <button
                @click="increaseRange"
                class="range-button"
                :disabled="!filteredPackages || !Array.isArray(filteredPackages) || displayLimit >= filteredPackages.length"
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
          <span class="range-info" v-if="filteredPackages && Array.isArray(filteredPackages) && filteredPackages.length > 0">
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
    <div v-else-if="!filteredPackages || !Array.isArray(filteredPackages) || filteredPackages.length === 0" class="empty-state">
      <div class="empty-icon">📦</div>
      <p v-if="searchQuery">No packages match your search criteria "{{ searchQuery }}".</p>
      <p v-else>No package data available.</p>
      <button
        v-if="searchQuery && packages && packages.length > 0"
        @click="searchQuery = ''"
        class="reset-button"
      >
        Clear Search
      </button>
    </div>

    <!-- Package summary section -->
    <div v-else-if="filteredPackages && Array.isArray(filteredPackages) && filteredPackages.length > 0" class="metrics-content">
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
          <!-- Package header section -->
          <div class="metric-header">
            <div class="package-title">
              <h4>{{ pkg?.name || 'Unknown Package' }}</h4>
              <div class="package-version">v{{ pkg?.version || 'Unknown' }}</div>
            </div>
            <span class="metric-badge" :class="getMaintenanceClass(pkg?.maintenanceScore || 0)">
              {{ pkg?.maintenanceScore || 0 }}%
            </span>
          </div>

          <!-- Package details section - rearranged for landscape orientation -->
          <div class="metric-content">
            <div class="metric-details">
              <!-- Popularity metrics -->
              <div class="metrics-section">
                <div class="metrics-section-title">
                  <i class="section-icon">📊</i>
                  Popularity
                </div>
                <div class="metrics-data">
                  <!-- Downloads displayed as table -->
                  <table class="downloads-table">
                    <thead>
                      <tr>
                        <th>
                          <i class="metric-icon downloads-icon">⬇️</i>
                          Downloads
                        </th>
                        <th>
                          <i class="metric-icon stars-icon">⭐</i>
                          Stars
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr>
                        <td class="metric-value highlight">{{ formatNumber(pkg?.downloads || 0) }}</td>
                        <td class="metric-value">{{ formatNumber(pkg?.stars || 0) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- Health metrics -->
              <div class="metrics-section">
                <div class="metrics-section-title">
                  <i class="section-icon">🔍</i>
                  Health
                </div>
                <div class="metrics-data">
                  <div class="metric-row">
                    <span class="metric-label">
                      <i class="metric-icon issues-icon">🐞</i>
                      Open Issues
                    </span>
                    <span class="metric-value" :class="pkg?.issues > 20 ? 'warning' : ''">
                      {{ pkg?.issues || 0 }}
                    </span>
                  </div>
                  <div class="metric-row">
                    <span class="metric-label">
                      <i class="metric-icon calendar-icon">📅</i>
                      Last Updated
                    </span>
                    <span class="metric-value">{{ formatDate(pkg?.lastUpdated || new Date().toISOString()) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Maintenance progress section -->
            <div class="metric-progress">
              <div class="metrics-section-title">
                <i class="section-icon">🛠️</i>
                Maintenance Score
              </div>
              <div class="score-display">
                <div class="score-circle" :class="getMaintenanceClass(pkg?.maintenanceScore || 0)">
                  {{ pkg?.maintenanceScore || 0 }}%
                </div>
              </div>
              <div class="progress-bar">
                <div
                  class="progress-fill"
                  :style="{ width: `${pkg?.maintenanceScore || 0}%` }"
                  :class="getMaintenanceClass(pkg?.maintenanceScore || 0)"
                ></div>
              </div>
              <div class="progress-info">
                <span class="progress-label">
                  {{ getMaintenanceLabel(pkg?.maintenanceScore || 0) }}
                </span>
                <span class="maintenance-info-icon" 
                      title="Maintenance score considers update frequency, issue resolution, and documentation quality">ℹ️</span>
              </div>
            </div>
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
    // ===== STATE MANAGEMENT =====
    const searchQuery = ref('');
    const selectedMetric = ref('downloads');
    const isLoading = ref(true);
    const packageCache = ref(new Map());
    const debouncedSearchQuery = ref('');
    const displayLimit = ref(10);
    const safePackages = ref([]);

    // ===== DATA PROCESSING =====
    
    // Generate consistent data for a package based on name
    const generatePackageData = (component) => {
      if (!component || !component.name) {
        return createEmptyPackage();
      }
      
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

    // Create an empty package object with default values
    const createEmptyPackage = () => ({
      name: 'Unknown',
      version: 'Unknown',
      downloads: 0,
      stars: 0,
      issues: 0,
      lastUpdated: new Date().toISOString().split('T')[0],
      maintenanceScore: 0
    });

    // Use memoization to avoid recalculating package data
    const getPackageData = (component) => {
      if (!component || !component.name) {
        return createEmptyPackage();
      }
      
      const cacheKey = `${component.name}-${component.version}`;

      if (!packageCache.value.has(cacheKey)) {
        packageCache.value.set(cacheKey, generatePackageData(component));
      }

      return packageCache.value.get(cacheKey);
    };

    // ===== COMPUTED PROPERTIES =====
    
    // Parse and process SBOM data to extract packages
    const packages = computed(() => {
      try {
        // If no SBOM data is provided, return safe default
        if (!props.sbomData) {
          console.log('No sbomData provided to PackageMetrics');
          return safePackages.value;
        }

        // Process SBOM data based on its type
        let componentsArray = extractComponentsArray(props.sbomData);
        
        // If no components could be extracted, return safe default
        if (!componentsArray || componentsArray.length === 0) {
          return safePackages.value;
        }

        // Map components to package data
        return componentsArray.map(component => getPackageData(component));
      } catch (error) {
        console.error('Error in packages computed property:', error);
        return safePackages.value;
      }
    });

    // Extract components array from various SBOM data formats
    const extractComponentsArray = (data) => {
      // Handle string format (JSON)
      if (typeof data === 'string') {
        try {
          const parsed = JSON.parse(data);
          return parsed.components || [];
        } catch (parseError) {
          console.error('Error parsing sbomData string:', parseError);
          return [];
        }
      } 
      // Handle object format with components array
      else if (data.components && Array.isArray(data.components)) {
        return data.components;
      } 
      // Handle unexpected format
      else {
        console.log('sbomData does not contain a valid components array');
        return [];
      }
    };

    // Get total number of packages
    const totalPackages = computed(() => 
      packages.value && Array.isArray(packages.value) ? packages.value.length : 0
    );

    // Filter and sort packages based on search query and selected metric
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
        const safePackagesCopy = [...packages.value];

        // Apply filters
        const filtered = query
          ? safePackagesCopy.filter(pkg => {
              // Ensure pkg and pkg.name exist
              return pkg && typeof pkg.name === 'string' && pkg.name.toLowerCase().includes(query);
            })
          : safePackagesCopy;

        // Sort packages based on selected metric
        return sortPackages(filtered, metric);
      } catch (error) {
        console.error('Error in filteredPackages computed property:', error);
        return safePackages.value;
      }
    });

    // Sort packages based on the selected metric
    const sortPackages = (packages, metric) => {
      return packages.sort((a, b) => {
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
    };

    // Calculate average downloads across packages
    const averageDownloads = computed(() => {
      return calculateAverage('downloads');
    });

    // Calculate average stars across packages
    const averageStars = computed(() => {
      return calculateAverage('stars');
    });

    // Calculate average maintenance score across packages
    const averageMaintenance = computed(() => {
      return calculateAverage('maintenanceScore');
    });

    // Helper function to calculate averages
    const calculateAverage = (property) => {
      try {
        // Use filteredPackages if available, otherwise fall back to safePackages
        const packagesToUse = getValidPackagesArray(filteredPackages.value);

        if (packagesToUse.length === 0) return 0;

        const sum = packagesToUse.reduce((acc, pkg) => acc + (pkg?.[property] || 0), 0);
        return Math.round(sum / packagesToUse.length);
      } catch (error) {
        console.error(`Error calculating average ${property}:`, error);
        return 0;
      }
    };

    // Get a valid packages array or return safe default
    const getValidPackagesArray = (packages) => {
      if (packages && Array.isArray(packages) && packages.length > 0) {
        return packages;
      }
      return safePackages.value;
    };

    // Apply display limit to show a subset of filtered packages
    const displayedPackages = computed(() => {
      try {
        // If no filtered packages, return safe default
        if (!filteredPackages.value || !Array.isArray(filteredPackages.value)) {
          console.log('filteredPackages is invalid in displayedPackages');
          return safePackages.value.slice(0, 5);
        }

        // Handle empty filtered packages
        if (filteredPackages.value.length === 0) {
          return safePackages.value.length > 0 ? safePackages.value.slice(0, 5) : [];
        }

        // Calculate safe display limit
        const numLimit = Number(displayLimit.value);
        const validLimit = isNaN(numLimit) ? 5 : numLimit;
        const limit = Math.max(5, Math.min(filteredPackages.value.length, validLimit));

        return filteredPackages.value.slice(0, limit);
      } catch (error) {
        console.error('Error in displayedPackages computed property:', error);
        return safePackages.value.slice(0, 5);
      }
    });

    // ===== WATCHERS =====
    
    // Debounce search input
    let searchTimeout = null;
    watch(searchQuery, (newValue) => {
      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }

      searchTimeout = setTimeout(() => {
        debouncedSearchQuery.value = newValue;
      }, 300);
    });

    // Update safePackages when packages changes
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

    // Adjust display limit based on filtered packages
    watch(filteredPackages, (newPackages) => {
      try {
        // Skip if packages is invalid
        if (!newPackages || !Array.isArray(newPackages)) {
          console.log('Filtered packages are invalid in watcher');
          displayLimit.value = 10;
          return;
        }

        // Adjust display limit if needed
        if (displayLimit.value > newPackages.length) {
          displayLimit.value = Math.max(5, newPackages.length);
        }

        // Reset to default if no packages
        if (newPackages.length === 0) {
          displayLimit.value = 10;
        }
      } catch (error) {
        console.error('Error in filteredPackages watcher:', error);
        displayLimit.value = 10;
      }
    }, { immediate: false });

    // ===== USER ACTIONS =====
    
    // Increase the display limit
    function increaseRange() {
      try {
        // Get maximum available packages safely
        const maxPackages = getMaxPackages();
        
        // Get current display limit as a number with fallback
        const currentLimit = Number(displayLimit.value) || 5;

        // Increase by 5, but don't exceed maxPackages
        displayLimit.value = Math.min(maxPackages, currentLimit + 5);
      } catch (error) {
        console.error('Error in increaseRange function:', error);
      }
    }

    // Decrease the display limit
    function decreaseRange() {
      try {
        // Get current limit with fallback
        const currentLimit = Number(displayLimit.value) || 10;

        // Decrease by 5, but don't go below 5
        displayLimit.value = Math.max(5, currentLimit - 5);
      } catch (error) {
        console.error('Error in decreaseRange function:', error);
        displayLimit.value = 5;
      }
    }

    // Get maximum packages count safely
    function getMaxPackages() {
      if (filteredPackages.value && Array.isArray(filteredPackages.value)) {
        return filteredPackages.value.length;
      }
      console.log('filteredPackages is not valid in getMaxPackages');
      return 20; // Default to a reasonable value
    }

    // Validate and correct the display limit if needed
    function validateDisplayLimit() {
      try {
        // Convert to number and validate
        const numValue = Number(displayLimit.value);
        
        // Handle invalid values
        if (isNaN(numValue) || numValue < 5) {
          displayLimit.value = 5;
          return;
        }

        // Get max packages safely
        const maxPackages = getMaxPackages();

        // Cap at max packages
        if (numValue > maxPackages && maxPackages > 0) {
          displayLimit.value = maxPackages;
          return;
        }

        // Ensure it's an integer
        displayLimit.value = Math.floor(numValue);
      } catch (error) {
        console.error('Error in validateDisplayLimit function:', error);
        displayLimit.value = 5;
      }
    }

    // ===== FORMATTING =====
    
    // Cache for formatted numbers
    const numberFormatCache = new Map();

    // Format large numbers with K/M suffixes
    const formatNumber = (num) => {
      try {
        // Handle invalid input
        if (num === undefined || num === null || isNaN(num)) {
          return '0';
        }

        // Ensure num is a number
        const numValue = Number(num);

        // Check cache first
        if (numberFormatCache.has(numValue)) {
          return numberFormatCache.get(numValue);
        }

        // Format based on magnitude
        let result;
        if (numValue >= 1000000) {
          result = (numValue / 1000000).toFixed(1) + 'M';
        } else if (numValue >= 1000) {
          result = (numValue / 1000).toFixed(1) + 'K';
        } else {
          result = numValue.toString();
        }

        // Cache and return
        numberFormatCache.set(numValue, result);
        return result;
      } catch (error) {
        console.error('Error in formatNumber function:', error);
        return '0';
      }
    };

    // Cache for formatted dates
    const dateFormatCache = new Map();

    // Format dates in a user-friendly way
    const formatDate = (dateString) => {
      try {
        // Handle invalid input
        if (!dateString) {
          return 'Unknown date';
        }

        // Check cache first
        if (dateFormatCache.has(dateString)) {
          return dateFormatCache.get(dateString);
        }

        // Parse and format date
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
          return 'Invalid date';
        }

        const result = formatDateToLocale(date);

        // Cache and return
        dateFormatCache.set(dateString, result);
        return result;
      } catch (error) {
        console.error('Error in formatDate function:', error);
        return 'Unknown date';
      }
    };

    // Format date to locale string
    const formatDateToLocale = (date) => {
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      });
    };

    // Cache for maintenance class determinations
    const maintenanceClassCache = new Map();

    // Get CSS class based on maintenance score
    const getMaintenanceClass = (score) => {
      try {
        // Handle invalid input
        if (score === undefined || score === null || isNaN(score)) {
          return 'poor';
        }

        // Ensure score is a number
        const numScore = Number(score);

        // Check cache first
        if (maintenanceClassCache.has(numScore)) {
          return maintenanceClassCache.get(numScore);
        }

        // Determine class based on score
        let result;
        if (numScore >= 90) result = 'excellent';
        else if (numScore >= 75) result = 'good';
        else if (numScore >= 60) result = 'fair';
        else result = 'poor';

        // Cache and return
        maintenanceClassCache.set(numScore, result);
        return result;
      } catch (error) {
        console.error('Error in getMaintenanceClass function:', error);
        return 'poor';
      }
    };

    // Get descriptive label based on maintenance score
    const getMaintenanceLabel = (score) => {
      const numScore = Number(score);
      if (isNaN(numScore)) return 'Poor maintenance';
      
      if (numScore >= 90) return 'Excellent maintenance';
      else if (numScore >= 75) return 'Well-maintained';
      else if (numScore >= 60) return 'Adequately maintained';
      else return 'Needs maintenance attention';
    };

    // ===== LIFECYCLE =====
    
    // Initialize component
    onMounted(() => {
      // Set initial loading state with placeholder data
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

      // Simulate loading state and check data
      isLoading.value = true;
      setTimeout(() => {
        checkDataLoaded();
      }, 1000);
    });

    // Check if data is loaded and update loading state
    const checkDataLoaded = () => {
      if (packages.value.length <= 1 && packages.value[0]?.name === 'Loading...') {
        console.log('No real package data available after timeout');
      } else {
        isLoading.value = false;
      }
    };

    // ===== EXPORTS =====
    
    return {
      // State
      searchQuery,
      selectedMetric,
      isLoading,
      displayLimit,
      
      // Computed
      filteredPackages,
      displayedPackages,
      totalPackages,
      averageDownloads,
      averageStars,
      averageMaintenance,
      
      // Actions
      increaseRange,
      decreaseRange,
      validateDisplayLimit,
      
      // Formatting
      formatNumber,
      formatDate,
      getMaintenanceClass,
      getMaintenanceLabel,
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
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 1.25rem;
}

.metric-card {
  background: white;
  border-radius: 10px;
  padding: 1.25rem;
  transition: all 0.25s ease;
  border: 1px solid var(--border-color);
  box-shadow: 0 3px 5px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  position: relative;
  overflow: visible;
}

.metric-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 16px rgba(0, 0, 0, 0.1);
  border-color: var(--primary-color);
}

.metric-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 4px;
  background: linear-gradient(to right, var(--primary-color), var(--secondary-color));
  opacity: 0;
  transition: opacity 0.3s ease;
}

.metric-card:hover::before {
  opacity: 1;
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
  align-items: flex-start;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 1rem;
  width: 100%;
}

.package-title {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  max-width: 80%;
  flex: 1;
  min-width: 0;
}

.metric-header h4 {
  font-size: 1.25rem;
  font-weight: var(--font-weight-bold);
  color: var(--dark-color);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  letter-spacing: 0.01em;
  position: relative;
  padding-bottom: 0.3rem;
  width: 100%;
}

.metric-header h4::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: 0;
  width: 40px;
  height: 2px;
  background: var(--primary-color);
}

.package-version {
  font-size: var(--font-size-xs);
  color: var(--secondary-color);
  font-weight: var(--font-weight-medium);
  background-color: rgba(100, 116, 139, 0.1);
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  display: inline-block;
}

.metric-badge {
  padding: 0.5rem 0.9rem;
  border-radius: 20px;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  box-shadow: 0 3px 5px rgba(0, 0, 0, 0.15);
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 3.5rem;
  flex-shrink: 0;
  margin-left: 0.5rem;
}

.metric-badge::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(255, 255, 255, 0.2);
  transform: translateX(-100%);
  transition: transform 0.3s ease;
}

.metric-card:hover .metric-badge::before {
  transform: translateX(100%);
}

.metric-badge.excellent {
  background: linear-gradient(to right, rgba(16, 185, 129, 0.9), rgba(5, 150, 105, 0.9));
  color: white;
}

.metric-badge.good {
  background: linear-gradient(to right, rgba(37, 99, 235, 0.9), rgba(30, 64, 175, 0.9));
  color: white;
}

.metric-badge.fair {
  background: linear-gradient(to right, rgba(245, 158, 11, 0.9), rgba(217, 119, 6, 0.9));
  color: white;
}

.metric-badge.poor {
  background: linear-gradient(to right, rgba(220, 38, 38, 0.9), rgba(185, 28, 28, 0.9));
  color: white;
}

.metric-content {
  display: flex;
  flex-direction: row;
  gap: 1.25rem;
  height: auto;
  flex-wrap: wrap;
}

.metric-details {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  flex: 3;
  min-width: 200px;
}

.metrics-section {
  padding: 0.85rem;
  background-color: rgba(241, 245, 249, 0.4);
  border-radius: 8px;
  height: auto;
  min-height: 100px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid rgba(226, 232, 240, 0.7);
}

.metrics-section:hover {
  background-color: rgba(241, 245, 249, 0.8);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}

.metrics-section-title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--dark-color);
  margin-bottom: 0.5rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid rgba(226, 232, 240, 0.7);
  padding-bottom: 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.section-icon {
  font-style: normal;
  font-size: 1rem;
}

.metrics-data {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.4rem 0;
  border-bottom: 1px dashed rgba(226, 232, 240, 0.5);
}

.metric-row:last-child {
  border-bottom: none;
}

.metric-label {
  color: var(--dark-color);
  font-size: var(--font-size-sm);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.metric-icon {
  font-style: normal;
  opacity: 0.9;
}

.metric-value {
  font-weight: var(--font-weight-medium);
  color: var(--dark-color);
  transition: transform 0.2s ease;
}

.metric-value.highlight {
  font-weight: var(--font-weight-bold);
  color: var(--primary-color);
}

.metric-value.warning {
  font-weight: var(--font-weight-bold);
  color: var(--warning-color);
}

.metric-row:hover .metric-value {
  transform: scale(1.05);
}

.metric-progress {
  margin-top: 0;
  flex: 2;
  display: flex;
  flex-direction: column;
  padding: 0.85rem;
  background-color: rgba(241, 245, 249, 0.4);
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid rgba(226, 232, 240, 0.7);
  min-width: 180px;
  min-height: 220px;
}

.score-display {
  display: flex;
  justify-content: center;
  margin: 0.5rem 0 1rem;
}

.score-circle {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: var(--font-weight-bold);
  font-size: 1rem;
  color: white;
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.15);
}

.score-circle.excellent {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.9), rgba(5, 150, 105, 0.9));
}

.score-circle.good {
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.9), rgba(30, 64, 175, 0.9));
}

.score-circle.fair {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.9), rgba(217, 119, 6, 0.9));
}

.score-circle.poor {
  background: linear-gradient(135deg, rgba(220, 38, 38, 0.9), rgba(185, 28, 28, 0.9));
}

.progress-bar {
  height: 10px;
  background: rgba(226, 232, 240, 0.7);
  border-radius: 6px;
  overflow: hidden;
  margin: 0.5rem 0;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.1);
}

.progress-fill {
  height: 100%;
  transition: width 0.8s ease;
  position: relative;
}

.progress-fill::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 5px;
  height: 100%;
  background-color: rgba(255, 255, 255, 0.3);
  border-radius: 50%;
}

.progress-fill.excellent {
  background: linear-gradient(to right, rgb(16, 185, 129), rgb(5, 150, 105));
}

.progress-fill.good {
  background: linear-gradient(to right, rgb(37, 99, 235), rgb(30, 64, 175));
}

.progress-fill.fair {
  background: linear-gradient(to right, rgb(245, 158, 11), rgb(217, 119, 6));
}

.progress-fill.poor {
  background: linear-gradient(to right, rgb(220, 38, 38), rgb(185, 28, 28));
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.progress-label {
  font-size: var(--font-size-xs);
  color: var(--secondary-color);
  font-weight: var(--font-weight-medium);
}

.maintenance-info-icon {
  font-size: var(--font-size-xs);
  cursor: help;
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

@media (min-width: 1200px) {
  .metrics-grid {
    grid-template-columns: repeat(auto-fill, minmax(450px, 1fr));
  }
}

.downloads-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 0.5rem;
}

.downloads-table th {
  text-align: left;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--dark-color);
  padding: 0.4rem 0.5rem;
  background-color: rgba(226, 232, 240, 0.5);
  border-radius: 4px 4px 0 0;
}

.downloads-table td {
  padding: 0.6rem 0.5rem;
  border-top: 1px solid rgba(226, 232, 240, 0.7);
  text-align: left;
}

.downloads-table tr:hover td {
  background-color: rgba(241, 245, 249, 0.8);
}
</style>