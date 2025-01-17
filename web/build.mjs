import { context, build } from 'esbuild';

const buildOne = async (path, options) => {
    const esbOpts = {
        outfile: path,
        bundle: true,
        ...options,
    };
    if (process.argv.includes('--watch')) {
        const ctx = await context({ ...esbOpts, logLevel: 'info' });
        await ctx.watch();
    } else {
        await build(esbOpts);
    }
};

const buildAll = () => {
    return Promise.all([
        buildOne('browser/script.js', {
            entryPoints: ['src/index.ts'],
            platform: 'browser',
            minify: true,
            sourcemap: 'inline',
            target: ['es6'],
        }),
        buildOne('dist/index.mjs', {
            entryPoints: ['src/index.ts'],
            platform: 'neutral',
            packages: 'external',
            sourcemap: true,
        }),
        buildOne('dist/index.js', {
            entryPoints: ['src/index.ts'],
            platform: 'node',
            target: ['node10.4'],
            sourcemap: true,
            packages: 'external',
        }),
    ]);
};

buildAll();
