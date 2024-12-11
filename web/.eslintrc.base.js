module.exports = {
    env: {
        node: true,
        es2023: true,
    },
    parser: '@typescript-eslint/parser',
    parserOptions: {
        sourceType: 'module',
        ecmaFeatures: {
            impliedStrict: true,
        },
        project: './tsconfig.json',
        tsconfigRootDir: process.env.TSCONFIG_ROOT_DIR || __dirname,
    },
    plugins: ['@typescript-eslint'],
    extends: ['eslint:recommended', 'plugin:@typescript-eslint/strict-type-checked', 'prettier'],
    rules: {
        complexity: ['error', 10],
        '@typescript-eslint/explicit-function-return-type': [
            'error',
            {
                allowExpressions: true,
            },
        ],
    },
};
