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
    <div class="graph-container" ref="graphContainer"></div>
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
import { ref, onMounted, watch, computed } from 'vue';

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

    const graphData = computed(() => {
      if (!props.sbomData || !props.sbomData.components) {
        return { nodes: [], edges: [] };
      }

      const nodes = [];
      const edges = [];
      const components = props.sbomData.components || [];
      const maxDepth = selectedDepth.value === 'all' ? Infinity : parseInt(selectedDepth.value);

      // Apply depth filtering
      const includedComponents = components.filter((_, index) => {
        // In a real implementation, you would determine the depth of each component
        // and filter based on maxDepth. For now, we'll include all components.
        return index < 10 * maxDepth; // Just for demonstration
      });

      // Create nodes
      includedComponents.forEach((component, index) => {
        nodes.push({
          id: index,
          label: component.name || `Component ${index}`,
          title: component.version ? `${component.name} v${component.version}` : component.name,
          group: component.type || 'default'
        });
      });

      // Create edges based on dependencies
      includedComponents.forEach((component, index) => {
        if (component.dependencies) {
          component.dependencies.forEach(dep => {
            const depIndex = components.findIndex(c => c.name === dep.name);
            if (depIndex !== -1 && depIndex < includedComponents.length) {
              edges.push({
                from: index,
                to: depIndex,
                arrows: 'to'
              });
            }
          });
        }
      });

      return { nodes, edges };
    });

    const createGraph = () => {
      if (!graphContainer.value) return;
      
      if (!window.vis) {
        console.error('vis-network not loaded');
        return;
      }

      const options = {
        nodes: {
          shape: 'dot',
          size: 16,
          font: {
            size: 12,
            color: '#000000'
          },
          borderWidth: 2
        },
        edges: {
          width: 2,
          smooth: {
            type: 'continuous'
          }
        },
        physics: {
          stabilization: false,
          barnesHut: {
            gravitationalConstant: -80000,
            springConstant: 0.001,
            springLength: 200
          }
        },
        layout: {
          improvedLayout: true
        }
      };

      if (network.value) {
        network.value.destroy();
      }

      network.value = new window.vis.Network(
        graphContainer.value,
        graphData.value,
        options
      );
    };

    // Load vis-network from CDN
    const loadVisNetwork = () => {
      return new Promise((resolve) => {
        if (window.vis) {
          resolve();
          return;
        }

        const script = document.createElement('script');
        script.src = 'https://unpkg.com/vis-network/standalone/umd/vis-network.min.js';
        script.onload = () => resolve();
        document.head.appendChild(script);
      });
    };

    watch([selectedDepth, graphData], () => {
      if (window.vis) {
        createGraph();
      }
    });

    onMounted(async () => {
      await loadVisNetwork();
      createGraph();
    });

    return {
      graphContainer,
      selectedDepth
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
</style> 