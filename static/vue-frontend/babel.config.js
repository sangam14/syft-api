module.exports = {
  presets: [
    ['@vue/cli-plugin-babel/preset', {
      useBuiltIns: 'entry',
      corejs: 3,
      targets: { browsers: ['> 1%', 'last 2 versions', 'not dead'] }
    }]
  ],

}