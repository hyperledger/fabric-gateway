module.exports = {
    env: {
        jest: true,
    },
    root: true,
    ignorePatterns: [
        'dist/',
    ],
    extends: [
        '.eslintrc.base',
    ],
    overrides: [
        {
            files: [
                '**/*.ts',
            ],
            plugins: [
                'jest',
                'eslint-plugin-tsdoc',
            ],
            extends: [
                'plugin:jest/recommended',
            ],
            rules: {
                'tsdoc/syntax': ['error'],
            },
        },
    ],
};
