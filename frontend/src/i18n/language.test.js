import { describe, expect, it } from 'vitest';
import { DEFAULT_LANGUAGE, getCopy, toLocale } from './language';

const dictionary = { 'pt-BR': 'português', en: 'inglês' };

describe('getCopy', () => {
  it.each([
    ['pt-BR', 'português'],
    ['en', 'inglês'],
    ['fr', 'português']
  ])('%s → %s', (language, expected) => {
    expect(getCopy(dictionary, language)).toBe(expected);
  });
});

describe('toLocale', () => {
  it.each([
    ['pt-BR', 'pt-BR'],
    ['en', 'en-US']
  ])('%s → %s', (language, expected) => {
    expect(toLocale(language)).toBe(expected);
  });
});

it('o idioma padrão é o português', () => {
  expect(DEFAULT_LANGUAGE).toBe('pt-BR');
});
