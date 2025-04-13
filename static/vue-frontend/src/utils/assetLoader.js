/**
 * Asset Loader Utility
 *
 * This utility provides functions to preload and manage external assets like
 * scripts, stylesheets, and images to improve application performance.
 */

// Track loaded assets to prevent duplicate loading
const loadedAssets = new Map();

/**
 * Preload a script with optimized loading strategy
 *
 * @param {string} src - Script URL to load
 * @param {Object} options - Loading options
 * @param {boolean} options.async - Whether to load the script asynchronously
 * @param {boolean} options.defer - Whether to defer script loading
 * @param {string} options.integrity - Integrity hash for the script
 * @param {Function} options.onLoad - Callback function when script loads
 * @returns {Promise} - Promise that resolves when the script is loaded
 */
export function preloadScript(src, options = {}) {
  const {
    async = true,
    defer = true,
    integrity = '',
    onLoad = null
  } = options;

  // Return existing promise if script is already loading
  if (loadedAssets.has(src)) {
    return loadedAssets.get(src);
  }

  const loadPromise = new Promise((resolve, reject) => {
    // Check if script is already in the document
    const existingScript = document.querySelector(`script[src="${src}"]`);
    if (existingScript) {
      resolve();
      return;
    }

    // Create and configure script element
    const script = document.createElement('script');
    script.src = src;
    script.async = async;
    script.defer = defer;

    if (integrity) {
      script.integrity = integrity;
      script.crossOrigin = 'anonymous';
    }

    // Set up event handlers
    script.onload = () => {
      if (onLoad && typeof onLoad === 'function') {
        onLoad();
      }
      resolve();
    };

    script.onerror = () => {
      loadedAssets.delete(src);
      reject(new Error(`Failed to load script: ${src}`));
    };

    // Add script to document
    document.head.appendChild(script);
  });

  // Store promise in cache
  loadedAssets.set(src, loadPromise);
  return loadPromise;
}

/**
 * Preload a stylesheet with optimized loading strategy
 *
 * @param {string} href - Stylesheet URL to load
 * @param {Object} options - Loading options
 * @param {string} options.media - Media attribute for the stylesheet
 * @param {string} options.integrity - Integrity hash for the stylesheet
 * @returns {Promise} - Promise that resolves when the stylesheet is loaded
 */
export function preloadStylesheet(href, options = {}) {
  const {
    media = 'all',
    integrity = ''
  } = options;

  // Return existing promise if stylesheet is already loading
  if (loadedAssets.has(href)) {
    return loadedAssets.get(href);
  }

  const loadPromise = new Promise((resolve, reject) => {
    // Check if stylesheet is already in the document
    const existingLink = document.querySelector(`link[href="${href}"]`);
    if (existingLink) {
      resolve();
      return;
    }

    // Create and configure link element
    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = href;
    link.media = media;

    if (integrity) {
      link.integrity = integrity;
      link.crossOrigin = 'anonymous';
    }

    // Set up event handlers
    link.onload = () => resolve();
    link.onerror = () => {
      loadedAssets.delete(href);
      reject(new Error(`Failed to load stylesheet: ${href}`));
    };

    // Add link to document
    document.head.appendChild(link);
  });

  // Store promise in cache
  loadedAssets.set(href, loadPromise);
  return loadPromise;
}

/**
 * Preload an image to cache it for later use
 *
 * @param {string} src - Image URL to preload
 * @returns {Promise} - Promise that resolves when the image is loaded
 */
export function preloadImage(src) {
  // Return existing promise if image is already loading
  if (loadedAssets.has(src)) {
    return loadedAssets.get(src);
  }

  const loadPromise = new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => {
      loadedAssets.delete(src);
      reject(new Error(`Failed to load image: ${src}`));
    };
    img.src = src;
  });

  // Store promise in cache
  loadedAssets.set(src, loadPromise);
  return loadPromise;
}

/**
 * Preload multiple assets in parallel
 *
 * @param {Array} assets - Array of asset objects to preload
 * @returns {Promise} - Promise that resolves when all assets are loaded
 */
export function preloadAssets(assets) {
  const promises = assets.map(asset => {
    switch (asset.type) {
      case 'script':
        return preloadScript(asset.src, asset.options);
      case 'stylesheet':
        return preloadStylesheet(asset.href, asset.options);
      case 'image':
        return preloadImage(asset.src);
      default:
        return Promise.reject(new Error(`Unknown asset type: ${asset.type}`));
    }
  });

  return Promise.all(promises);
}

/**
 * Check if an asset is already loaded
 *
 * @param {string} src - Asset URL to check
 * @returns {boolean} - Whether the asset is loaded
 */
export function isAssetLoaded(src) {
  return loadedAssets.has(src) && loadedAssets.get(src).status === 'fulfilled';
}

/**
 * Clear the asset cache
 */
export function clearAssetCache() {
  loadedAssets.clear();
}

export default {
  preloadScript,
  preloadStylesheet,
  preloadImage,
  preloadAssets,
  isAssetLoaded,
  clearAssetCache
};
