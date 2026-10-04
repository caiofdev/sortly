import { describe, expect, it } from 'vitest';
import { SortlyError } from '../services/sortlyGateway';
import { describeError } from './describeError';
import feedbackCopy from './feedbackCopy';
import { getCopy, toLocale } from './language';

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
  'SETTINGS_NOT_SAVED'
];

describe('language', () => {
  it('getCopy usa o idioma pedido ou o padrão', () => {
    expect(getCopy(feedbackCopy, 'en')).toBe(en);
    expect(getCopy(feedbackCopy, 'fr')).toBe(pt);
  });

  it('toLocale', () => {
    expect(toLocale('pt-BR')).toBe('pt-BR');
    expect(toLocale('en')).toBe('en-US');
  });
});

describe('códigos de erro do backend (ADR 0004)', () => {
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

describe('describeError', () => {
  it('traduz o código no idioma atual', () => {
    const error = new SortlyError('NOTHING_TO_UNDO', 'NOTHING_TO_UNDO');
    expect(describeError(error, pt, 'padrão')).toBe('Nenhuma separação recente para desfazer.');
    expect(describeError(error, en, 'default')).toBe('No recent organization to undo.');
  });

  it('UNEXPECTED vindo do backend usa o texto padrão', () => {
    expect(describeError(new SortlyError('UNEXPECTED', 'UNEXPECTED'), pt, 'padrão')).toBe('padrão');
  });

  it('texto livre, backend indisponível e código sem tradução usam o texto padrão', () => {
    expect(describeError(new SortlyError('UNEXPECTED', 'TypeError: x'), pt, 'padrão')).toBe(
      'padrão'
    );
    expect(describeError(new SortlyError('UNEXPECTED', 'Backend indisponível'), pt, 'padrão')).toBe(
      'padrão'
    );
    expect(describeError(new SortlyError('NOVO_CODIGO', 'NOVO_CODIGO'), pt, 'padrão')).toBe(
      'padrão'
    );
    expect(describeError(null, pt, 'padrão')).toBe('padrão');
  });
});

describe('mensagens de sucesso', () => {
  const base = {
    movedFiles: 3,
    sourceFolderPath: 'C:\\origem',
    destinationFolderPath: 'C:\\destino',
    processedFiles: 5,
    ignoredWithoutExtension: 1,
    ignoredFolders: 1
  };

  it('sem falhas nem inalterados: mesma mensagem da versão anterior', () => {
    expect(pt.organizeSuccess({ ...base, failedFiles: 0, unchangedFiles: 0 })).toBe(
      'Organização concluida: 3 arquivo(s) movido(s). Origem: C:\\origem. Destino: C:\\destino. Processados: 5. Ignorados sem extensão: 1. Pastas ignoradas: 1.'
    );
  });

  it('com 1 falha e 1 inalterado: acrescenta os trechos', () => {
    const result = { ...base, failedFiles: 1, unchangedFiles: 1 };
    expect(pt.organizeSuccess(result)).toMatch(/Já estavam no lugar: 1\. Falhas ao mover: 1\.$/);
    expect(en.organizeSuccess(result)).toMatch(/Already in place: 1\. Failed to move: 1\.$/);
  });

  it('desfazer com e sem falhas', () => {
    const result = { restoredFiles: 2, renamedOnRestore: 0, skippedMissing: 0, failedFiles: 0 };
    expect(pt.undoSuccess(result)).toBe(
      'Desfazer concluido: 2 arquivo(s) restaurado(s). Renomeados na restauração: 0. Não encontrados: 0.'
    );
    expect(en.undoSuccess({ ...result, failedFiles: 1 })).toMatch(/Failed to restore: 1\.$/);
  });
});
