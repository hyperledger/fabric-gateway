import tseslint from 'typescript-eslint';
import base from '../../node/eslint.config.base.mjs';

export default tseslint.config(...base, {
    languageOptions: {
        parserOptions: {
            tsconfigRootDir: import.meta.dirname,
        },
    },
});
