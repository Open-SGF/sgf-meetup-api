const CAMEL_CASE_PATTERN = '+([a-z])*([a-z0-9])*([A-Z]*([a-z0-9]))';
const PASCAL_CASE_PATTERN = '*([A-Z]*([a-z0-9]))';

module.exports = {
	root: true,
	env: {
		node: true,
	},
	extends: [
		'plugin:@typescript-eslint/recommended',
		'plugin:prettier/recommended',
		'plugin:import/errors',
		'plugin:import/warnings',
		'plugin:import/typescript',
	],
	parser: '@typescript-eslint/parser',
	parserOptions: {
		ecmaVersion: 2020,
		sourceType: 'module',
		project: 'tsconfig.json',
	},
	plugins: ['@typescript-eslint/eslint-plugin', 'check-file'],
	rules: {
		'prettier/prettier': 'error',
		'no-console': 'warn',
		'import/no-default-export': 'error',
		'check-file/filename-naming-convention': [
			'error',
			{
				'**/*.{jsx,tsx,ts,js}': `@(${CAMEL_CASE_PATTERN}|${PASCAL_CASE_PATTERN})`,
			},
		],
		'check-file/folder-naming-convention': [
			'error',
			{
				'{lambdas,scripts}/**/': `@(${CAMEL_CASE_PATTERN}|${PASCAL_CASE_PATTERN})`,
			},
		],
	},
	ignorePatterns: ['cdk.out'],
	overrides: [
		{
			files: ['**/*.ts'],
			rules: {
				'@typescript-eslint/no-explicit-any': 'warn',
				'@typescript-eslint/no-floating-promises': 'error',
				'@typescript-eslint/no-misused-promises': 'error',
				'@typescript-eslint/no-unused-vars': 'error',
				'@typescript-eslint/prefer-optional-chain': 'error',
			},
		},
		{
			files: ['lambdas/**/*.ts'],
			parserOptions: {
				project: 'lambdas/tsconfig.json',
			},
		},
		{
			files: ['.eslintrc.js'],
			parser: 'espree',
			rules: {
				'prettier/prettier': 'error',
			},
		},
	],
};
