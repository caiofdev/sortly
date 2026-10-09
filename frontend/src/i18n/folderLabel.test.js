import { describe, expect, it } from 'vitest';
import organizerCopy from './organizerCopy';
import { folderLabel, folderPathLabel } from './folderLabel';

const pt = organizerCopy['pt-BR'];
const en = organizerCopy.en;

describe('folderPathLabel', () => {
  it.each([
    ['', true, pt.previewRoot],
    ['images', true, 'Imagens'],
    ['images\\2026-09-14', true, 'Imagens\\2026-09-14'],
    ['images/2026-09-14', true, 'Imagens/2026-09-14'],
    ['images\\2026-09-14', false, 'images\\2026-09-14']
  ])('pasta "%s", categorias = %s → "%s"', (path, categoryFolders, expected) => {
    expect(folderPathLabel(path, categoryFolders, pt)).toBe(expected);
  });
});

describe('folderLabel', () => {
  it.each([
    ['', false, pt.previewRoot],
    ['', true, pt.previewRoot],
    ['audio', false, 'audio'],
    ['images', true, 'Imagens'],
    ['other', true, 'Outros'],
    ['desconhecida', true, 'desconhecida']
  ])('pasta "%s", categorias = %s → "%s"', (name, categoryFolders, expected) => {
    expect(folderLabel(name, categoryFolders, pt)).toBe(expected);
  });

  it('as sete categorias têm rótulo em PT e EN', () => {
    const keys = ['images', 'documents', 'archives', 'installers', 'videos', 'audio', 'other'];
    for (const key of keys) {
      expect(pt.categories[key]).toBeTruthy();
      expect(en.categories[key]).toBeTruthy();
    }
    expect(folderLabel('audio', true, en)).toBe('Audio');
  });
});
