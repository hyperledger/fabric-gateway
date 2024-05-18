import js from '@eslint/js';
import prettier from 'eslint-config-prettier';
import tseslint from 'typescript-eslint';

export default tseslint.config(js.configs.recommended, ...tseslint.configs.strictTypeChecked, prettier, {
    languageOptions: {
        ecmaVersion: 2023,
        sourceType: 'module',
        parserOptions: {
            project: 'tsconfig.json',
            tsconfigRootDir: import.meta.dirname,
        },
    },
    rules: {
        complexity: ['error', 10],
        '@typescript-eslint/explicit-function-return-type': [
            'error',
            {
                allowExpressions: true,
            },
        ],
    },
});
