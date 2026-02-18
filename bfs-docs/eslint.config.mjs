import nextConfig from 'eslint-config-next';

const eslintConfig = [
  ...nextConfig,
  {
    ignores: ['.source/**'],
  },
];

export default eslintConfig;