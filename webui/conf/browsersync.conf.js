const conf = require('./gulp.conf');
const proxy = require('http-proxy-middleware');

const apiProxy = proxy('/api', {
  target: 'http://localhost:8080',
  changeOrigin: true
});

const healthProxy = proxy('/health', {
  target: 'http://localhost:8080',
  changeOrigin: true
});

module.exports = function () {
  return {
    server: {
      baseDir: [
        conf.paths.tmp,
        conf.paths.src
      ],
      middleware: [
        apiProxy,
        healthProxy
      ]
    },
    open: false
  };
};
