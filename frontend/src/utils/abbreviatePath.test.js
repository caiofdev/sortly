import { describe, expect, it } from 'vitest';
import { MAX_PATH_LENGTH, abbreviatePath } from './abbreviatePath';

describe('abbreviatePath', () => {
  it('72 caracteres: mantém o caminho inteiro', () => {
    const path = 'a'.repeat(MAX_PATH_LENGTH);
    expect(abbreviatePath(path)).toBe(path);
  });

  it('73 caracteres: mostra 32 do começo, "..." e 32 do fim', () => {
    const path = `${'i'.repeat(32)}MEIO-CORTADO-X${'f'.repeat(27)}`;
    expect(path).toHaveLength(MAX_PATH_LENGTH + 1);
    expect(abbreviatePath(path)).toBe(`${'i'.repeat(32)}...${path.slice(-32)}`);
  });

  it('vazio continua vazio', () => {
    expect(abbreviatePath('')).toBe('');
  });
});
