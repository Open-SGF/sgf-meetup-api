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
	plugins: ['@typescript-eslint/eslint-plugin'],
	rules: {
		'prettier/prettier': 'error',
		'no-console': 'warn',
		'import/no-default-export': 'error',
		'@typescript-eslint/no-explicit-any': 'warn',
		'@typescript-eslint/no-floating-promises': 'error',
		'@typescript-eslint/no-misused-promises': 'error',
		'@typescript-eslint/no-unused-vars': 'error',
		'@typescript-eslint/prefer-optional-chain': 'error',
	},
	ignorePatterns: ['cdk.out', '.eslintrc.js'],
	overrides: [
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
			}
		},
	],
};
