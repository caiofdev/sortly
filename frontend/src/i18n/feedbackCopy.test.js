import { describe, expect, it } from 'vitest';
import feedbackCopy from './feedbackCopy';

const pt = feedbackCopy['pt-BR'];
const en = feedbackCopy.en;
const BACKEND_CODES = [
  'INVALID_SOURCE',
  'INVALID_DESTINATION',
  'NO_CRITERIA',
  'NOTHING_TO_UNDO',
  'DROPPED_INVALID',
  'DROPPED_MISSING',
  'DROPPED_UNSUPPORTED',
  'RECORD_NOT_SAVED',
  'LAST_CRITERION',
  'UNKNOWN_CRITERION',
  'INVALID_LANGUAGE',
  'INVALID_THEME',
  'INVALID_DUPLICATES',
  'DESTINATION_NOT_FOUND',
  'SETTINGS_NOT_SAVED'
];
const NOTICE_TEXTS = ['sourceRequired', 'recovered', 'unexpectedError'];

describe('códigos de erro do backend (ADR 0004)', () => {
  it.each(NOTICE_TEXTS)('%s existe em PT e EN', (key) => {
    expect(pt[key]).toBeTruthy();
    expect(en[key]).toBeTruthy();
  });

  it.each(BACKEND_CODES)('%s tem tradução em PT e EN', (code) => {
    expect(pt.errors[code]).toBeTruthy();
    expect(en.errors[code]).toBeTruthy();
  });

  it('textos em PT iguais aos da versão 1.0', () => {
    expect(pt.errors.INVALID_SOURCE).toBe('Pasta inválida.');
    expect(pt.errors.INVALID_DESTINATION).toBe('Pasta de destino inválida.');
    expect(pt.errors.NO_CRITERIA).toBe('Selecione ao menos um criterio de organizacao.');
    expect(pt.errors.NOTHING_TO_UNDO).toBe('Nenhuma separação recente para desfazer.');
  });
});

describe('arquivos duplicados (#82)', () => {
  const organized = (overrides) => ({
    movedFiles: 1,
    destinationFolderPath: 'D',
    unchangedFiles: 0,
    ignoredWithoutExtension: 0,
    failedFiles: 0,
    skippedDuplicates: 0,
    replacedFiles: 0,
    ...overrides
  });

  it.each([
    [
      { skippedDuplicates: 1 },
      '1 já existia no destino e ficou na origem.',
      '1 file was already in the destination and stayed in the source.'
    ],
    [
      { skippedDuplicates: 2 },
      '2 já existiam no destino e ficaram na origem.',
      '2 files were already in the destination and stayed in the source.'
    ],
    [
      { replacedFiles: 1 },
      '1 arquivo antigo foi substituído e fica guardado para o desfazer.',
      '1 old file was replaced and kept for undo.'
    ],
    [
      { replacedFiles: 2 },
      '2 arquivos antigos foram substituídos e ficam guardados para o desfazer.',
      '2 old files were replaced and kept for undo.'
    ]
  ])('%o', (overrides, ptText, enText) => {
    expect(pt.organizeDone(organized(overrides)).text).toContain(ptText);
    expect(en.organizeDone(organized(overrides)).text).toContain(enText);
  });

  it('desfazer conta os substituídos que voltaram', () => {
    const r = {
      restoredFiles: 1,
      renamedOnRestore: 0,
      skippedMissing: 0,
      failedFiles: 0,
      restoredReplaced: 1
    };
    expect(pt.undoDone(r).text).toContain('1 substituído voltou para o lugar.');
    expect(en.undoDone({ ...r, restoredReplaced: 2 }).text).toContain(
      '2 replaced files are back in place.'
    );
  });
});

describe('organização cancelada (#78)', () => {
  it.each([
    [0, 'Organização cancelada', 'Organizing cancelled'],
    [1, 'Interrompido: 1 arquivo movido', 'Stopped: 1 file moved'],
    [2, 'Interrompido: 2 arquivos movidos', 'Stopped: 2 files moved']
  ])('%i movido(s): "%s"', (movedFiles, ptTitle, enTitle) => {
    expect(pt.organizeCanceled({ movedFiles }).title).toBe(ptTitle);
    expect(en.organizeCanceled({ movedFiles }).title).toBe(enTitle);
  });

  // Regressão (#78): o protótipo dizia "nada foi alterado" mesmo com arquivos já movidos.
  it('com arquivos movidos, o texto fala em desfazer', () => {
    expect(pt.organizeCanceled({ movedFiles: 5 }).text).toBe(
      'Você pode desfazer o que já foi movido.'
    );
    expect(pt.organizeCanceled({ movedFiles: 0 }).text).toBe(
      'Nada foi alterado na pasta de origem.'
    );
  });
});

describe('avisos de organizar e desfazer (#75)', () => {
  const result = (overrides = {}) => ({
    movedFiles: 3,
    destinationFolderPath: 'C:\\destino',
    unchangedFiles: 0,
    ignoredWithoutExtension: 0,
    failedFiles: 0,
    ...overrides
  });

  it.each([
    [result(), 'Pronto! 3 arquivos organizados', 'Done! 3 files organized'],
    [result({ movedFiles: 1 }), 'Pronto! 1 arquivo organizado', 'Done! 1 file organized'],
    [result({ movedFiles: 0 }), 'Nenhum arquivo para organizar', 'No files to organize'],
    [
      result({ movedFiles: 2, failedFiles: 1 }),
      '2 arquivos organizados, 1 com falha',
      '2 files organized, 1 failed'
    ]
  ])('título de %o', (r, ptTitle, enTitle) => {
    expect(pt.organizeDone(r).title).toBe(ptTitle);
    expect(en.organizeDone(r).title).toBe(enTitle);
  });

  it('o texto só traz os detalhes maiores que zero', () => {
    expect(pt.organizeDone(result()).text).toBe('Na pasta C:\\destino.');
    expect(pt.organizeDone(result({ movedFiles: 0 })).text).toBe('');
  });

  it('com todos os detalhes, em PT e EN', () => {
    const r = result({ unchangedFiles: 2, ignoredWithoutExtension: 1, failedFiles: 1 });
    expect(pt.organizeDone(r).text).toBe(
      'Na pasta C:\\destino. 2 já estavam no lugar. 1 sem extensão ficou na origem. 1 não pôde ser movido.'
    );
    expect(en.organizeDone(r).text).toBe(
      'In C:\\destino. 2 were already in place. 1 file without extension stayed in the source. 1 file could not be moved.'
    );
  });

  const undone = (overrides = {}) => ({
    restoredFiles: 2,
    renamedOnRestore: 0,
    skippedMissing: 0,
    failedFiles: 0,
    ...overrides
  });

  it.each([
    [undone(), 'Organização desfeita', 'Organizing undone'],
    [undone({ failedFiles: 1 }), 'Organização desfeita, 1 com falha', 'Organizing undone, 1 failed']
  ])('título do desfazer %o', (r, ptTitle, enTitle) => {
    expect(pt.undoDone(r).title).toBe(ptTitle);
    expect(en.undoDone(r).title).toBe(enTitle);
  });

  it('texto do desfazer: sem detalhes e com todos', () => {
    expect(pt.undoDone(undone({ restoredFiles: 1 })).text).toBe('1 arquivo voltou para a origem.');
    const r = undone({ renamedOnRestore: 1, skippedMissing: 2, failedFiles: 1 });
    expect(pt.undoDone(r).text).toBe(
      '2 arquivos voltaram para a origem. 1 foi renomeado para não sobrescrever outro. 2 não foram encontrados. Desfaça de novo para tentar restaurar o resto.'
    );
    expect(en.undoDone(r).text).toBe(
      '2 files are back in the source folder. 1 was renamed to avoid overwriting another. 2 were not found. Undo again to try restoring the rest.'
    );
  });

  it('origem solta mostra o caminho como texto', () => {
    expect(pt.sourceDropped('C:\\solta')).toEqual({ title: 'Origem definida', text: 'C:\\solta' });
    expect(en.sourceDropped('C:\\solta').title).toBe('Source set');
  });
});
