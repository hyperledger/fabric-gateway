process.env.TSCONFIG_ROOT_DIR = __dirname;

module.exports = {
    root: true,
    ignorePatterns: [
        'dist/',
    ],
    extends: [
        '../../node/.eslintrc.base.js',
    ],
};
