/**
 * @param {string|undefined} key
 * @param {Options} options
 * @param {Compilation} compilation
 * @returns {{lint: Linter, report: Reporter, threads: number}}
 */
export default function linter(
  key: string | undefined,
  options: Options,
  compilation: Compilation
): {
  lint: Linter;
  report: Reporter;
  threads: number;
};
export type ESLint = import('eslint').ESLint;
export type Formatter = import('eslint').ESLint.Formatter;
export type LintResult = import('eslint').ESLint.LintResult;
export type Compiler = import('webpack').Compiler;
export type Compilation = import('webpack').Compilation;
export type Options = import('./options').Options;
export type FormatterFunction = import('./options').FormatterFunction;
export type GenerateReport = (compilation: Compilation) => Promise<void>;
export type Report = {
  errors?: ESLintError;
  warnings?: ESLintError;
  generateReportAsset?: GenerateReport;
};
export type Reporter = () => Promise<Report>;
export type Linter = (files: string | string[]) => void;
export type LintResultMap = {
  [files: string]: import('eslint').ESLint.LintResult;
};
import ESLintError from './ESLintError';
