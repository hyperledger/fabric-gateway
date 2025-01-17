module.exports = {
    root: true,
    env: {
        jest: true,
    },
    ignorePatterns: ['*/**', '*.js', '*.ts', '!src/**/*.ts', '*.mjs'],
    plugins: ['jest', 'eslint-plugin-tsdoc'],
    extends: ['.eslintrc.base', 'plugin:jest/recommended'],
    rules: {
        'tsdoc/syntax': ['error'],
    },
};
