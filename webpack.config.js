const path = require('path');
const TerserPlugin = require('terser-webpack-plugin');

module.exports = {
    mode: 'production', // Ensures minification by default in production mode
    entry: './templates/components/index.js', // Entry point of your application
    output: {
        path: path.resolve("/Users/home/Documents/Code/Go/modulacms/public/js"),
        filename: 'bundle.js', // Output file
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: /node_modules/,
            },
        ],
    },
    optimization: {
        minimize: true, // Enables minification
        minimizer: [
            new TerserPlugin({
                terserOptions: {
                    compress: true,
                    mangle: true,
                },
            }),
        ],
    },
    resolve: {
        extensions: ['.js'], // Resolve these extensions
    },
};

