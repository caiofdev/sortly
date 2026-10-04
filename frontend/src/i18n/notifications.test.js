import { describe, expect, it } from 'vitest';
import feedbackCopy from './feedbackCopy';
import { notificationMessage, toNotificationItems } from './notifications';

const pt = feedbackCopy['pt-BR'];
const en = feedbackCopy.en;

describe('notificationMessage', () => {
  it.each([
    [{ code: 'RECOVERED_LAST_ORGANIZATION' }, pt.recoveredLastOrganization],
    [
      { code: 'SOURCE_DROPPED', path: 'C:\\solta' },
      'Origem definida por arrastar e soltar. C:\\solta'
    ],
    [{ code: 'SOURCE_REQUIRED' }, 'Selecione a pasta de origem antes de organizar.'],
    [{ code: 'NOTHING_TO_UNDO', action: 'undo' }, 'Nenhuma separação recente para desfazer.']
  ])('%o', (notification, expected) => {
    expect(notificationMessage(notification, pt)).toBe(expected);
  });

  it('resultado de organizar e desfazer usa os dados da notificação', () => {
    const organize = {
      movedFiles: 1,
      sourceFolderPath: 'C:\\o',
      destinationFolderPath: 'C:\\d',
      processedFiles: 1,
      ignoredWithoutExtension: 0,
      ignoredFolders: 0,
      failedFiles: 0,
      unchangedFiles: 0
    };
    expect(notificationMessage({ code: 'ORGANIZE_DONE', organize }, en)).toMatch(
      /^Organization complete: 1 file\(s\) moved\./
    );
    const undo = { restoredFiles: 2, renamedOnRestore: 0, skippedMissing: 0, failedFiles: 0 };
    expect(notificationMessage({ code: 'UNDO_DONE', undo }, pt)).toMatch(
      /^Desfazer concluido: 2 arquivo\(s\) restaurado\(s\)/
    );
  });

  it.each([
    ['selectSource', pt.sourceSelectError],
    ['selectDestination', pt.destinationSelectError],
    ['drop', pt.droppedPathUnexpectedError],
    ['organize', pt.organizeUnexpectedError],
    ['undo', pt.undoUnexpectedError],
    ['settings', pt.settingsSaveError],
    ['desconhecida', pt.unexpectedError]
  ])('UNEXPECTED na ação %s usa o texto padrão da ação', (action, expected) => {
    expect(notificationMessage({ code: 'UNEXPECTED', action }, pt)).toBe(expected);
  });
});

describe('toNotificationItems', () => {
  it('converte para o formato do NotificationsCenter, no idioma pedido', () => {
    const at = new Date(2026, 2, 5, 9, 7).toISOString();
    const [item] = toNotificationItems(
      [{ id: 7, kind: 'error', code: 'SOURCE_REQUIRED', action: 'organize', at }],
      'en'
    );
    expect(item).toEqual({
      id: 7,
      type: 'error',
      message: 'Select a source folder before organizing.',
      time: expect.stringMatching(/09:07/)
    });
  });

  it('lista vazia continua vazia', () => {
    expect(toNotificationItems([], 'pt-BR')).toEqual([]);
  });
});
