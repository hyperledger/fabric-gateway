module.exports = {
  "roots": [
    "<rootDir>/src"
  ],
  "preset": "ts-jest",
  "testEnvironment": "node",
  "collectCoverage": true,
  "collectCoverageFrom": [
//    '**/*.ts',
    "src/**/*.{js,jsx,ts,tsx}",
    "!src/protos/*.{ts,js}",
    "!**/node_modules/"
  ],
  "coverageProvider": "v8",
  "testMatch": [
    "**/?(*.)+(spec|test).+(ts|tsx|js)"
  ],
  "transform": {
//    "^.+\\.(ts|tsx)$": "ts-jest"
    "\\.ts": "ts-jest"
  },
}
