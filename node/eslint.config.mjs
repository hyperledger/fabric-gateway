import { FlatCompat } from '@eslint/eslintrc';
import jest from 'eslint-plugin-jest';
import { defineConfig } from 'eslint/config';
import base from './eslint.config.base.mjs';

const compat = new FlatCompat({ baseDirectory: import.meta.dirname });

export default defineConfig(...base, jest.configs['flat/recommended']);
