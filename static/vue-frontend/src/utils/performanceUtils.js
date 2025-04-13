/**
 * Performance Utilities for Vue Components
 * 
 * This module provides utilities to optimize Vue component performance.
 */

/**
 * Creates a debounced version of a function that delays invoking the function
 * until after the specified wait time has elapsed since the last time it was invoked.
 * 
 * @param {Function} func - The function to debounce
 * @param {number} wait - The number of milliseconds to delay
 * @param {boolean} immediate - Whether to invoke the function immediately
 * @returns {Function} - The debounced function
 */
export function debounce(func, wait = 300, immediate = false) {
  let timeout;
  
  return function executedFunction(...args) {
    const context = this;
    
    const later = function() {
      timeout = null;
      if (!immediate) func.apply(context, args);
    };
    
    const callNow = immediate && !timeout;
    
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
    
    if (callNow) func.apply(context, args);
  };
}

/**
 * Creates a throttled version of a function that only invokes the function
 * at most once per every specified wait milliseconds.
 * 
 * @param {Function} func - The function to throttle
 * @param {number} wait - The number of milliseconds to throttle invocations to
 * @returns {Function} - The throttled function
 */
export function throttle(func, wait = 300) {
  let lastFunc;
  let lastRan;
  
  return function executedFunction(...args) {
    const context = this;
    
    if (!lastRan) {
      func.apply(context, args);
      lastRan = Date.now();
    } else {
      clearTimeout(lastFunc);
      
      lastFunc = setTimeout(function() {
        if ((Date.now() - lastRan) >= wait) {
          func.apply(context, args);
          lastRan = Date.now();
        }
      }, wait - (Date.now() - lastRan));
    }
  };
}

/**
 * Memoizes a function to cache its results based on the arguments provided.
 * 
 * @param {Function} func - The function to memoize
 * @returns {Function} - The memoized function
 */
export function memoize(func) {
  const cache = new Map();
  
  return function memoized(...args) {
    const key = JSON.stringify(args);
    
    if (cache.has(key)) {
      return cache.get(key);
    }
    
    const result = func.apply(this, args);
    cache.set(key, result);
    
    return result;
  };
}

/**
 * Measures the execution time of a function and logs it to the console.
 * 
 * @param {Function} func - The function to measure
 * @param {string} name - The name of the function for logging
 * @returns {Function} - The wrapped function
 */
export function measurePerformance(func, name = 'Function') {
  return function measured(...args) {
    const start = performance.now();
    const result = func.apply(this, args);
    const end = performance.now();
    
    console.log(`${name} execution time: ${(end - start).toFixed(2)}ms`);
    
    return result;
  };
}

/**
 * Defers the execution of a function until the browser's next idle period.
 * Falls back to setTimeout if requestIdleCallback is not available.
 * 
 * @param {Function} func - The function to defer
 * @param {Object} options - Options for requestIdleCallback
 * @returns {Function} - The deferred function
 */
export function deferExecution(func, options = { timeout: 1000 }) {
  return function deferred(...args) {
    const context = this;
    
    if (window.requestIdleCallback) {
      window.requestIdleCallback(() => func.apply(context, args), options);
    } else {
      setTimeout(() => func.apply(context, args), 1);
    }
  };
}

/**
 * Creates a function that will only execute once the component is visible in the viewport.
 * 
 * @param {Function} func - The function to execute when visible
 * @param {Object} options - IntersectionObserver options
 * @returns {Function} - A function that sets up the visibility detection
 */
export function executeWhenVisible(func, options = { threshold: 0.1 }) {
  return function setupVisibility(element) {
    if (!element || !window.IntersectionObserver) {
      // Fall back to immediate execution if IntersectionObserver is not supported
      func();
      return;
    }
    
    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) {
        func();
        observer.disconnect();
      }
    }, options);
    
    observer.observe(element);
    
    // Return a cleanup function
    return () => observer.disconnect();
  };
}

/**
 * Batches multiple DOM updates together to reduce layout thrashing.
 * 
 * @param {Function} updateFunc - The function that performs DOM updates
 * @returns {Function} - The batched update function
 */
export function batchDomUpdates(updateFunc) {
  let scheduled = false;
  let updates = [];
  
  return function scheduleBatch(...args) {
    updates.push(args);
    
    if (!scheduled) {
      scheduled = true;
      
      window.requestAnimationFrame(() => {
        const batchedUpdates = updates;
        updates = [];
        scheduled = false;
        
        batchedUpdates.forEach(updateArgs => {
          updateFunc.apply(this, updateArgs);
        });
      });
    }
  };
}

export default {
  debounce,
  throttle,
  memoize,
  measurePerformance,
  deferExecution,
  executeWhenVisible,
  batchDomUpdates
};
