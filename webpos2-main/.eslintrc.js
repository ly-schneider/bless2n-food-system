// .eslintrc.js
module.exports = {
    parser: '@typescript-eslint/parser',
    parserOptions: {
    ecmaVersion: 2020,
    sourceType: 'module',
    ecmaFeatures: {
        jsx: true,
    },
    },
    settings: {
    react: {
        version: 'detect',
    },
    },
    plugins: ['react', '@typescript-eslint'],
    extends: [
    'next/core-web-vitals',
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended',
    'plugin:react-hooks/recommended'
    ],
    rules: {
    // Customize your rules
    'react/prop-types': 'off', // Since using TypeScript for props
    '@typescript-eslint/no-unused-vars': ['error'],
    '@typescript-eslint/no-explicit-any': 'warn',
    // ... other rules ...
    },
}
