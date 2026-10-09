import { describe, expect, it } from 'vitest';
import feedbackCopy from './feedbackCopy';
import { notificationText, toNotificationItems } from './notifications';

const pt = feedbackCopy['pt-BR'];
const en = feedbackCopy.en;

describe('notificationText', () => {
  it.each([
    [{ code: 'RECOVERED_LAST_ORGANIZATION' }, pt.recovered],
    [
      { code: 'SOURCE_DROPPED', path: 'C:\\solta' },
      { title: 'Origem definida', text: 'C:\\solta' }
    ],
    [
      { code: 'SOURCE_REQUIRED' },
      { title: 'Selecione a pasta de origem antes de organizar.', text: '' }
    ],
    [
      { code: 'NOTHING_TO_UNDO', action: 'undo' },
      { title: 'Nenhuma separação recente para desfazer.', text: '' }
    ]
  ])('%o', (notification, expected) => {
    expect(notificationText(notification, pt)).toEqual(expected);
  });

  it('resultado de organizar e desfazer usa os dados da notificação', () => {
    const organize = {
      movedFiles: 1,
      destinationFolderPath: 'C:\\d',
      ignoredWithoutExtension: 0,
      failedFiles: 0,
      unchangedFiles: 0
    };
    expect(notificationText({ code: 'ORGANIZE_DONE', organize }, en)).toEqual({
      title: 'Done! 1 file organized',
      text: 'In C:\\d.'
    });
    const undo = { restoredFiles: 2, renamedOnRestore: 0, skippedMissing: 0, failedFiles: 0 };
    expect(notificationText({ code: 'UNDO_DONE', undo }, pt)).toEqual({
      title: 'Organização desfeita',
      text: '2 arquivos voltaram para a origem.'
    });
    expect(
      notificationText({ code: 'ORGANIZE_CANCELED', organize: { movedFiles: 4 } }, pt).title
    ).toBe('Interrompido: 4 arquivos movidos');
  });

  it.each([
    ['selectSource', pt.sourceSelectError],
    ['selectDestination', pt.destinationSelectError],
    ['drop', pt.droppedPathUnexpectedError],
    ['organize', pt.organizeUnexpectedError],
    ['undo', pt.undoUnexpectedError],
    ['settings', pt.settingsSaveError],
    ['preview', pt.previewError],
    ['desconhecida', pt.unexpectedError]
  ])('UNEXPECTED na ação %s usa o texto padrão da ação', (action, expected) => {
    expect(notificationText({ code: 'UNEXPECTED', action }, pt)).toEqual({
      title: expected,
      text: ''
    });
  });
});

describe('toNotificationItems', () => {
  it('converte para o formato do painel e dos toasts, no idioma pedido', () => {
    const at = new Date(2026, 2, 5, 9, 7).toISOString();
    const [item] = toNotificationItems(
      [{ id: 7, kind: 'error', code: 'SOURCE_REQUIRED', action: 'organize', at }],
      'en'
    );
    expect(item).toEqual({
      id: 7,
      kind: 'error',
      title: 'Select a source folder before organizing.',
      text: '',
      time: expect.stringMatching(/09:07/)
    });
  });

  it('lista vazia continua vazia', () => {
    expect(toNotificationItems([], 'pt-BR')).toEqual([]);
  });
});
