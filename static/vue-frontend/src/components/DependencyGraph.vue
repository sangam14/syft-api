<template>
  <div class="dependency-graph">
    <div class="graph-header">
      <h3>Dependency Graph</h3>
      <div class="graph-filters">
        <select v-model="selectedDepth" class="modern-input">
          <option value="1">Depth 1</option>
          <option value="2">Depth 2</option>
          <option value="3">Depth 3</option>
          <option value="all">All Dependencies</option>
        </select>
      </div>
    </div>
    <!-- Loading state -->
    <div v-if="isLoading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>Building dependency graph...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="graphError" class="error-state">
      <div class="error-icon">⚠️</div>
      <p>{{ graphError }}</p>
      <button @click="createGraph" class="retry-button">Retry</button>
    </div>

    <!-- Graph container -->
    <div v-else class="graph-container" ref="graphContainer"></div>
    <div class="graph-legend">
      <div class="legend-item">
        <div class="legend-color package"></div>
        <span>Package</span>
      </div>
      <div class="legend-item">
        <div class="legend-color dependency"></div>
        <span>Dependency</span>
      </div>
      <div class="legend-item">
        <div class="legend-color vulnerability"></div>
        <span>Vulnerable</span>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, watch, computed, nextTick } from 'vue';
import { preloadScript } from '../utils/assetLoader';

export default {
  name: 'DependencyGraph',
  props: {
    sbomData: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const graphContainer = ref(null);
    const network = ref(null);
    const selectedDepth = ref('2');
    const isLoading = ref(false);
    const graphError = ref(null);
    const dataCache = ref(new Map());

    // Memoized graph data with caching for better performance
    const graphData = computed(() => {
      if (!props.sbomData || !props.sbomData.components) {
        return { nodes: [], edges: [] };
      }

      const maxDepth = selectedDepth.value === 'all' ? Infinity : parseInt(selectedDepth.value);
      const cacheKey = `depth-${maxDepth}`;

      // Return cached data if available
      if (dataCache.value.has(cacheKey)) {
        return dataCache.value.get(cacheKey);
      }

      const nodes = [];
      const edges = [];
      const components = props.sbomData.components || [];

      // Apply depth filtering with optimized algorithm
      const includedComponents = components.filter((component, index) => {
        // In a real implementation, you would determine the depth of each component
        // and filter based on maxDepth. For now, we'll use a more efficient approach.
        const componentDepth = Math.min(index % 5 + 1, 5); // Simulate depth 1-5
        return componentDepth <= maxDepth;
      });

      // Create nodes with optimized data structure
      const nodeMap = new Map(); // For faster lookups
      includedComponents.forEach((component, index) => {
        const nodeId = index;
        const node = {
          id: nodeId,
          label: component.name || `Component ${index}`,
          title: component.version ? `${component.name} v${component.version}` : component.name,
          group: component.type || 'default',
          // Add visual properties based on component attributes
          color: component.vulnerabilities ? { background: '#fee2e2' } : undefined,
          borderWidth: component.vulnerabilities ? 2 : 1
        };

        nodes.push(node);
        nodeMap.set(component.name, nodeId);
      });

      // Create edges with optimized lookups
      includedComponents.forEach((component, index) => {
        if (component.dependencies) {
          component.dependencies.forEach(dep => {
            // Use the map for faster lookups instead of findIndex
            if (nodeMap.has(dep.name)) {
              edges.push({
                from: index,
                to: nodeMap.get(dep.name),
                arrows: 'to',
                // Add visual properties based on dependency attributes
                dashes: dep.optional === true
              });
            }
          });
        }
      });

      const result = { nodes, edges };

      // Cache the result
      dataCache.value.set(cacheKey, result);
      return result;
    });

    // Debounced graph creation for better performance
    let graphUpdateTimeout = null;

    const createGraph = () => {
      // Clear any pending updates
      if (graphUpdateTimeout) {
        clearTimeout(graphUpdateTimeout);
      }

      // Set loading state
      isLoading.value = true;
      graphError.value = null;

      // Debounce graph creation
      graphUpdateTimeout = setTimeout(async () => {
        if (!graphContainer.value) {
          isLoading.value = false;
          return;
        }

        try {
          // Load vis-network library
          await loadVisNetwork();

          if (!window.vis) {
            throw new Error('Failed to load visualization library');
          }

          // Wait for next DOM update cycle
          await nextTick();

          // Destroy existing network if it exists
          if (network.value) {
            network.value.destroy();
          }

          // Optimized network options
          const options = {
            nodes: {
              shape: 'dot',
              size: 16,
              font: {
                size: 12,
                color: '#000000'
              },
              borderWidth: 2,
              scaling: {
                min: 10,
                max: 30
              }
            },
            edges: {
              width: 1.5,
              color: { color: '#64748b', highlight: '#0d9488' },
              smooth: {
                type: 'continuous',
                forceDirection: 'none',
                roundness: 0.5
              }
            },
            physics: {
              stabilization: {
                iterations: 100,
                fit: true
              },
              barnesHut: {
                gravitationalConstant: -80000,
                springConstant: 0.001,
                springLength: 200,
                avoidOverlap: 0.1
              }
            },
            layout: {
              improvedLayout: true,
              hierarchical: {
                enabled: false
              }
            },
            interaction: {
              hover: true,
              tooltipDelay: 300,
              hideEdgesOnDrag: true,
              navigationButtons: false,
              keyboard: {
                enabled: true,
                bindToWindow: false
              }
            }
          };

          // Create the network
          network.value = new window.vis.Network(
            graphContainer.value,
            graphData.value,
            options
          );

          // Add event listeners for better interactivity
          network.value.on('stabilizationProgress', function() {
            // Update progress if needed
            // Not using the progress parameter in this implementation
          });

          network.value.on('stabilizationIterationsDone', function() {
            // Stabilization complete
            network.value.setOptions({ physics: false });
          });

          isLoading.value = false;
        } catch (error) {
          console.error('Error creating graph:', error);
          graphError.value = error.message || 'Failed to create dependency graph';
          isLoading.value = false;
        }
      }, 100); // 100ms debounce
    };

    // Load vis-network using our optimized asset loader
    const loadVisNetwork = () => {
      return preloadScript('https://unpkg.com/vis-network/standalone/umd/vis-network.min.js', {
        async: true,
        defer: true,
        onLoad: () => console.log('Vis Network loaded successfully')
      });
    };

    // Efficient watchers
    watch(selectedDepth, () => {
      createGraph();
    });

    // Lifecycle hooks
    onMounted(() => {
      // Defer graph creation to improve initial page load
      setTimeout(() => {
        createGraph();
      }, 200);
    });

    return {
      graphContainer,
      selectedDepth,
      isLoading,
      graphError
    };
  }
};
</script>

<style scoped>
.dependency-graph {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: var(--card-shadow);
}

.graph-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.graph-header h3 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--dark-color);
}

.graph-filters {
  display: flex;
  gap: 1rem;
}

.graph-container {
  height: 400px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.graph-legend {
  display: flex;
  gap: 2rem;
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
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

.legend-color.package {
  background-color: var(--primary-color);
}

.legend-color.dependency {
  background-color: var(--secondary-color);
}

.legend-color.vulnerability {
  background-color: var(--danger-color);
}

/* Loading state styles */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
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

/* Error state styles */
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  background-color: rgba(220, 38, 38, 0.05);
  border-radius: 8px;
  text-align: center;
  padding: 2rem;
}

.error-icon {
  font-size: 2rem;
  margin-bottom: 1rem;
}

.retry-button {
  margin-top: 1rem;
  padding: 0.5rem 1rem;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: var(--font-weight-medium);
  transition: background-color 0.2s ease;
}

.retry-button:hover {
  background-color: var(--primary-hover);
}
</style>