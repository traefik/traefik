/* eslint angular/module-getter:0 */
const gulp = require('gulp');
const gutil = require('gulp-util');

const webpack = require('webpack');
const webpackConf = require('../conf/webpack.conf');
const webpackDistConf = require('../conf/webpack-dist.conf');
const browsersync = require('browser-sync');

gulp.task('webpack:dev', done => {
  webpackWrapper(false, webpackConf, done);
});

gulp.task('webpack:watch', done => {
  webpackWrapper(true, webpackConf, done);
});

gulp.task('webpack:dist', done => {
  webpackWrapper(false, webpackDistConf, done);
});

function webpackWrapper(watch, conf, done) {
  const webpackBundler = webpack(conf);

  const webpackChangeHandler = (err, stats) => {
    if (err) {
      conf.errorHandler('Webpack')(err);
    }
    gutil.log(stats.toString({
      colors: true,
      chunks: false,
      hash: false,
      version: false
    }));
    if (done) {
      done();
      done = null;
    } else {
      browsersync.reload();
    }
  };

  if (watch) {
    webpackBundler.watch(200, webpackChangeHandler);
  } else {
    webpackBundler.run(webpackChangeHandler);
  }
}
