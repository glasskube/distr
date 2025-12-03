/**
 * @see https://prettier.io/docs/en/configuration.html
 * @type {import("prettier").Config}
 */
const config = {
  bracketSameLine: true,
  bracketSpacing: false,
  plugins: ['prettier-plugin-organize-imports'],
  printWidth: 120,
  semi: true,
  singleQuote: true,
  tabWidth: 2,
  trailingComma: 'es5',
};

export default config;
