import type { Config } from 'jest';

const config: Config = {
    roots: ['<rootDir>/tests'],
    preset: 'ts-jest',
    testEnvironment: '<rootDir>/tests/customEnv.ts',
    collectCoverage: true,
    collectCoverageFrom: ['src/**/*.[jt]s?(x)', '!**/*.d.ts'],
    coverageProvider: 'v8',
    testMatch: ['**/?(*.)+(spec|test).[jt]s?(x)'],
    verbose: true,
    workerThreads: true,
};

export default config;
