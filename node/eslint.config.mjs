import jest from 'eslint-plugin-jest';
import tseslint from 'typescript-eslint';
import base from './eslint.config.base.mjs';
import { FlatCompat } from '@eslint/eslintrc';

const compat = new FlatCompat({ baseDirectory: import.meta.dirname });

export default tseslint.config(...base, jest.configs['flat/recommended'], ...compat.plugins('eslint-plugin-tsdoc'), {
    rules: {
        'tsdoc/syntax': ['error'],
    },
});
