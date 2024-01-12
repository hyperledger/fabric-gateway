process.env.TSCONFIG_ROOT_DIR = __dirname;

module.exports = {
    root: true,
    ignorePatterns: ['*/**', '*.js', '*.ts', '!src/**/*.ts'],
    extends: ['../../node/.eslintrc.base.js'],
};
