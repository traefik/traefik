const { resolve } = require('path');
const AngularCompilerPlugin = require('@ngtools/webpack').AngularCompilerPlugin;
const webpack = require('webpack');
const webpackMerge = require('webpack-merge');
const compression = require('compression-webpack-plugin');
const html = require('html-webpack-plugin');
const copy = require('copy-webpack-plugin');
const extract = require('extract-text-webpack-plugin');
const circular = require('circular-dependency-plugin');
const portfinder = require('portfinder');
const nodeModules = resolve(__dirname, 'node_modules');
const entryPoints = ["inline", "polyfills", "sw-register", "styles", "vendor", "app"];

module.exports = function (options, webpackOptions) {
  options = options || {};

  let config = {};

  config = webpackMerge({}, config, {
    entry: getEntry(options),
    resolve: {
      extensions: ['.ts', '.js', '.json'],
      modules: ['node_modules', nodeModules]
    },
    resolveLoader: {
      modules: [nodeModules, 'node_modules']
    },
    output: {
      path: root('../static')
    },
    module: {
      rules: [
        { test: /\.html$/, loader: 'html-loader', options: { minimize: true, removeAttributeQuotes: false, caseSensitive: true, customAttrSurround: [ [/#/, /(?:)/], [/\*/, /(?:)/], [/\[?\(?/, /(?:)/] ], customAttrAssign: [ /\)?\]?=/ ] } },
        { test: /\.json$/, loader: 'json-loader' },
        { test: /\.(jp?g|png|gif)$/, loader: 'file-loader', options: { hash: 'sha512', digest: 'hex', name: 'images/[hash].[ext]' } },
        { test: /\.(eot|woff2?|svg|ttf|otf)([\?]?.*)$/, loader: 'file-loader', options: { hash: 'sha512', digest: 'hex', name: 'fonts/[hash].[ext]' } }
      ]
    },
    plugins: [
      new copy([{ context: './src/assets/public', from: '**/*' }])
    ]
  });

  config = webpackMerge({}, config, {
    output: {
      path: root('../static'),
      filename: '[name].bundle.js',
      chunkFilename: '[id].chunk.js'
    },
    plugins: [
      new html({
        template: root('src/index.html'),
        output: root('dist'),
        chunksSortMode: sort = (left, right) => {
          let leftIndex = entryPoints.indexOf(left.names[0]);
          let rightindex = entryPoints.indexOf(right.names[0]);
          if (leftIndex > rightindex) {
            return 1;
          } else if (leftIndex < rightindex) {
            return -1;
          } else {
            return 0;
          }
        }
      })
    ],
    devServer: {
      historyApiFallback: true,
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true
        },
        '/health': {
          target: 'http://localhost:8080',
          changeOrigin: true
        }
      },
      overlay: true,
      port: 8080,
      open: true,
      hot: false,
      inline: true,
      stats: {
        errors: true,
        errorDetails: true,
        depth: false,
        chunkOrigins: false,
        chunkModules: false,
        chunks: false,
        children: false,
        cacheAssets: false,
        cached: false,
        assets: false,
        modules: false,
        hash: false,
        reasons: false,
        source: false,
        timings: true,
        version: false,
        warnings: false
      },
      watchOptions: {
        aggregateTimeout: 300,
        poll: 1000
      }
    }
  });

  if (webpackOptions.p) {
    config = webpackMerge({}, config, getProductionPlugins());
    config = webpackMerge({}, config, getProdStylesConfig());
  } else {
    config = webpackMerge({}, config, getDevelopmentConfig());
    config = webpackMerge({}, config, getDevStylesConfig());
  }

  config = webpackMerge({}, config, {
    module: {
      rules: [{ test: /(?:\.ngfactory\.js|\.ngstyle\.js|\.ts)$/, loader: '@ngtools/webpack' }]
    },
    plugins: [
      new AngularCompilerPlugin({ tsConfigPath: root('src/tsconfig.json'), entryModule: 'src/app.module#AppModule' })
    ]
  });

  if (options.serve) {
    return portfinder.getPortPromise().then(port => {
      config.devServer.port = port;
      return config;
    });
  } else {
    return Promise.resolve(config);
  }
}

function root(path) {
  return resolve(__dirname, path);
}

function getEntry(options) {
  if (options.aot) {
    return { app: root('src/main.ts') };
  } else {
    return { app: root('src/main.ts'), polyfills: root('src/polyfills.ts') };
  }
}

function getDevelopmentConfig() {
  return {
    devtool: 'inline-source-map',
    module: {
      rules: [
        { enforce: 'pre', test: /\.js$/, loader: 'source-map-loader', exclude: [nodeModules] }
      ]
    },
    plugins: [
      new webpack.NoEmitOnErrorsPlugin(),
      new webpack.NamedModulesPlugin(),
      new webpack.optimize.CommonsChunkPlugin({
        minChunks: Infinity,
        name: 'inline'
      }),
      new webpack.optimize.CommonsChunkPlugin({
        name: 'vendor',
        chunks: ['app'],
        minChunks: module => {
          return module.resource && module.resource.startsWith(nodeModules)
        }
      })
    ]
  };
}

function getProductionPlugins() {
  return {
    plugins: [
      new compression({ asset: "[path].gz[query]", algorithm: "gzip", test: /\.js$|\.html$/, threshold: 10240, minRatio: 0.8 })
    ]
  };
}

function getDevStylesConfig() {
  return {
    module: {
      rules: [
        { test: /\.css$/, use: ['style-loader', 'css-loader'], exclude: [root('src')] },
        { test: /\.css$/, use: ['to-string-loader', 'css-loader'], exclude: [root('src/styles')] },
        { test: /\.scss$|\.sass$/, use: ['style-loader', 'css-loader', 'sass-loader'], include: [root('src/styles') ] },
        { test: /\.scss$|\.sass$/, use: ['to-string-loader', 'css-loader', 'sass-loader'], exclude: [root('src/styles')] },
      ]
    }
  };
}

function getProdStylesConfig() {
  return {
    plugins: [
      new extract('css/[hash].css')
    ],
    module: {
      rules: [
        { test: /\.css$/, use: extract.extract({ fallback: 'style-loader', use: 'css-loader' }), include: [root('src/styles')] },
        { test: /\.css$/, use: ['to-string-loader', 'css-loader'], exclude: [root('src/styles')] },
        { test: /\.scss$|\.sass$/, loader: extract.extract({ fallback: 'style-loader', use: ['css-loader', 'sass-loader'] }), exclude: [root('src/app')] },
        { test: /\.scss$|\.sass$/, use: ['to-string-loader', 'css-loader', 'sass-loader'], exclude: [root('src/styles')] },
      ]
    }
  };
}
