module.exports = {
    preset: 'ts-jest',
    transform: {
        '^.+\\.ts?$': 'ts-jest',
        '^.+\\.js?$': 'ts-jest'
    },
    transformIgnorePatterns: ['node_modules/'],
    testEnvironment: 'node',
    testRegex: '/tests/.*\\.(test|spec)?\\.(ts|tsx)$',
    moduleNameMapper: { '^@/(.*)': '<rootDir>/src/$1' },
    moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node']
};