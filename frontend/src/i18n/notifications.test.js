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
    [pt, 0, 'Desfazer cancelado', 'Nada voltou para a origem; você ainda pode desfazer.'],
    [pt, 1, 'Desfazer interrompido: 1 arquivo voltou', 'Desfaça de novo para devolver o resto.'],
    [pt, 2, 'Desfazer interrompido: 2 arquivos voltaram', 'Desfaça de novo para devolver o resto.'],
    [en, 0, 'Undo cancelled', 'Nothing went back to the source; you can still undo.'],
    [en, 1, 'Undo stopped: 1 file is back', 'Undo again to return the rest.'],
    [en, 2, 'Undo stopped: 2 files are back', 'Undo again to return the rest.']
  ])('desfazer cancelado com %#: %s voltaram → "%s"', (copy, restoredFiles, title, text) => {
    expect(notificationText({ code: 'UNDO_CANCELED', undo: { restoredFiles } }, copy)).toEqual({
      title,
      text
    });
  });

  it.each([
    ['selectSource', pt.sourceSelectError],
    ['selectDestination', pt.destinationSelectError],
    ['drop', pt.droppedPathUnexpectedError],
    ['organize', pt.organizeUnexpectedError],
    ['undo', pt.undoUnexpectedError],
    ['settings', pt.settingsSaveError],
    ['preview', pt.previewError],
    ['openDestination', pt.openDestinationError],
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
